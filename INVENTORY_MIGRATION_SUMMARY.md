# Inventory Items Migration to Database-Driven System

## Overview
Successfully migrated the inventory items system from hardcoded values to a database-driven approach. This change allows for dynamic management of inventory items through the database while maintaining backward compatibility.

## Changes Made

### 1. Database Schema Updates
**File**: `/workspace/go/migrate/internal/migrations/000095_migrate_inventory_items_to_database.up.sql`
- Added new columns to `inventory_items` table:
  - `inventory_item_id` (INTEGER, unique): Maps to the old enum values
  - `rarity_tier` (VARCHAR): Item rarity ("Common", "Uncommon", "Epic", "Mythic", "Not Droppable")  
  - `is_capture_type` (BOOLEAN): Indicates if item can capture challenges
  - `item_type` (VARCHAR): Type of item ("consumable", "passive", "equippable")
  - `equipment_slot` (VARCHAR): Equipment slot for equippable items
- Created unique index on `inventory_item_id`
- Seeded database with all 20 existing hardcoded items
- **Cleared all existing owned inventory items as requested**

**File**: `/workspace/go/migrate/internal/migrations/000095_migrate_inventory_items_to_database.down.sql`
- Rollback migration to remove new columns and data

### 2. Go Backend Updates

#### New Model
**File**: `/workspace/go/pkg/models/inventory_item.go`
- Created new `InventoryItem` model matching the database schema
- Includes proper GORM tags for database mapping

#### Database Interface Extensions
**File**: `/workspace/go/pkg/db/interfaces.go`
- Extended `InventoryItemHandle` interface with new methods:
  - `GetAllInventoryItems(ctx context.Context) ([]models.InventoryItem, error)`
  - `FindInventoryItemByID(ctx context.Context, inventoryItemID int) (*models.InventoryItem, error)`

#### Database Implementation
**File**: `/workspace/go/pkg/db/inventory_item.go`
- Implemented new methods to fetch inventory items from database
- Added proper error handling and context support

#### Quartermaster Updates
**File**: `/workspace/go/sonar/internal/quartermaster/client.go`
- Updated `GetInventoryItems()` to fetch from database with fallback to hardcoded items
- Updated `FindItemForItemID()` to prioritize database lookup
- Maintains backward compatibility with existing `PreDefinedItems`

### 3. Frontend Type Updates

#### TypeScript Interface Updates
**File**: `/workspace/js/packages/types/src/inventoryItem.ts`
- Updated `InventoryItem` type:
  - Changed `id` from `ItemType` enum to `number`
  - Added optional fields: `isCaptureType`, `itemType`, `equipmentSlot`
- Updated item reference arrays to use numeric IDs instead of enum values
- Maintained `ItemType` enum for backward compatibility

## Database Seeded Items

The migration seeds 20 inventory items with IDs 1-20:

1. **Cipher of the Laughing Monkey** (Uncommon, Consumable)
2. **Golden Telescope** (Uncommon, Consumable) 
3. **Flawed Ruby** (Uncommon, Consumable, Capture Type)
4. **Ruby** (Epic, Consumable, Capture Type)
5. **Brilliant Ruby** (Mythic, Consumable, Capture Type)
6. **Cortez's Cutlass** (Not Droppable, Equippable, Right Hand)
7. **Rusted Musket** (Common, Consumable)
8. **Gold Coin** (Common, Passive)
9. **Dagger** (Epic, Equippable, Left Hand)
10. **Damage** (Not Droppable, Passive)
11. **Entseed** (Not Droppable, Passive)
12. **Ale** (Uncommon, Consumable)
13. **Witchflame** (Not Droppable, Passive)
14. **Wicked Spellbook** (Not Droppable, Equippable, Left Hand)
15. **The Compass of Peace** (Not Droppable, Equippable, Neck)
16. **Pirate's Tricorn Hat** (Uncommon, Equippable, Head)
17. **Captain's Coat** (Epic, Equippable, Chest)
18. **Seafarer's Boots** (Common, Equippable, Feet)
19. **Enchanted Ring of Fortune** (Mythic, Equippable, Ring)
20. **Leather Sailing Gloves** (Common, Equippable, Gloves)

## Migration Execution

To apply these changes:

1. **Start PostgreSQL database**:
   ```bash
   # If Docker is available:
   make deps
   # Or manually start PostgreSQL with credentials from deps.docker-compose.yml
   ```

2. **Build and run migration**:
   ```bash
   cd go/migrate
   go build -o migrate cmd/migrate/main.go
   DB_PASSWORD=test_password ./migrate -config-name=local -config-path=.
   ```

3. **Verify migration**:
   ```sql
   SELECT COUNT(*) FROM inventory_items; -- Should return 20
   SELECT COUNT(*) FROM owned_inventory_items; -- Should return 0 (all cleared)
   ```

## API Impact

- **GET `/sonar/items`**: Now returns database items instead of hardcoded items
- Response format remains unchanged for frontend compatibility
- All existing item IDs (1-20) are preserved
- Performance improved with proper database indexing

## Benefits

1. **Dynamic Management**: Items can now be added/modified via database without code changes
2. **Scalability**: Easy to add new items through database inserts
3. **Data Integrity**: Proper foreign key relationships and constraints
4. **Backward Compatibility**: Existing code continues to work with numeric IDs
5. **Fallback Safety**: Graceful fallback to hardcoded items if database fails

## Testing Recommendations

1. Verify API endpoint `/sonar/items` returns correct data
2. Test item usage functionality with database-driven items  
3. Confirm equipment system works with new item data
4. Validate owned inventory items are properly cleared
5. Test fallback behavior when database is unavailable

## Future Considerations

- Consider adding admin interface for managing inventory items
- Implement item versioning for balance changes
- Add validation rules for item properties
- Consider caching frequently accessed items
- Plan for item localization/internationalization