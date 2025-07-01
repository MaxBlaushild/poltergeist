# Inventory Items Migration to Database

## Overview
This migration transforms the inventory system from using hardcoded items to being fully database-driven. All previously hardcoded inventory items are now stored in the database and can be managed dynamically.

## Changes Made

### 1. Database Migration
- **File**: `go/migrate/internal/migrations/000096_update_inventory_items_structure_and_seed.up.sql`
- **Purpose**: Updates the existing `inventory_items` table structure and seeds it with all hardcoded items
- **Changes**:
  - Changed `id` from UUID to SERIAL (auto-incrementing integer)
  - Added `rarity_tier` column
  - Added `is_capture_type` column  
  - Added `item_type` column
  - Added `equipment_slot` column
  - Seeded table with all 20 hardcoded inventory items

- **File**: `go/migrate/internal/migrations/000096_update_inventory_items_structure_and_seed.down.sql`
- **Purpose**: Rollback migration to revert changes

### 2. Database Model
- **File**: `go/pkg/models/inventory_item.go`
- **Purpose**: GORM model for database operations
- **Features**:
  - Proper GORM column mappings
  - JSON serialization tags
  - Table name specification

### 3. Database Handler
- **File**: `go/pkg/db/inventory_item.go`
- **Purpose**: Database operations for inventory items
- **Methods**:
  - `FindAll()` - Get all inventory items
  - `FindByID()` - Get item by ID
  - `Create()` - Create new item
  - `Update()` - Update existing item
  - `Delete()` - Delete item

### 4. Database Interface Updates
- **File**: `go/pkg/db/interfaces.go`
- **Purpose**: Added `InventoryItemHandle` interface to the main `DbClient`
- **Changes**:
  - Added `InventoryItem()` method to `DbClient` interface
  - Added `InventoryItemHandle` interface definition

### 5. Database Client Updates  
- **File**: `go/pkg/db/client.go`
- **Purpose**: Integrated inventory item handler into the main database client
- **Changes**:
  - Added `inventoryItemHandler` to client struct
  - Added handler initialization in constructor
  - Added `InventoryItem()` accessor method

### 6. Quartermaster Service Updates
- **File**: `go/sonar/internal/quartermaster/client.go`
- **Purpose**: Updated to use database instead of hardcoded items
- **Changes**:
  - `GetInventoryItems()` now queries database with fallback to hardcoded items
  - `FindItemForItemID()` now queries database first, then falls back to hardcoded
  - `getRandomItem()` now uses database items for random selection
  - Added proper model conversion between database and quartermaster models

### 7. Bug Fix
- **File**: `go/pkg/models/user_stats.go`
- **Purpose**: Fixed duplicate constant declaration
- **Change**: Removed duplicate `StatPointsPerLevel` constant (kept in `user_level.go`)

## Migration Data
The migration seeds the database with all 20 existing hardcoded inventory items:

1. Cipher of the Laughing Monkey (Uncommon, Consumable)
2. Golden Telescope (Uncommon, Consumable)
3. Flawed Ruby (Uncommon, Consumable, Capture Type)
4. Ruby (Epic, Consumable, Capture Type)
5. Brilliant Ruby (Mythic, Consumable, Capture Type)
6. Cortez's Cutlass (Not Droppable, Equippable, Right Hand)
7. Rusted Musket (Common, Consumable)
8. Gold Coin (Common, Passive)
9. Dagger (Epic, Equippable, Left Hand)
10. Damage (Not Droppable, Passive)
11. Entseed (Not Droppable, Passive)
12. Ale (Uncommon, Consumable)
13. Witchflame (Not Droppable, Passive)
14. Wicked Spellbook (Not Droppable, Equippable, Left Hand)
15. The Compass of Peace (Not Droppable, Equippable, Neck)
16. Pirate's Tricorn Hat (Uncommon, Equippable, Head)
17. Captain's Coat (Epic, Equippable, Chest)
18. Seafarer's Boots (Common, Equippable, Feet)
19. Enchanted Ring of Fortune (Mythic, Equippable, Ring)
20. Leather Sailing Gloves (Common, Equippable, Gloves)

## Backward Compatibility
- The system maintains backward compatibility by falling back to hardcoded items if database operations fail
- All item IDs remain the same (1-20) to maintain compatibility with existing owned inventory items
- The API response format remains unchanged

## Benefits
1. **Dynamic Management**: Items can now be added, modified, or removed without code changes
2. **Database Consistency**: Item data is now properly normalized and stored in the database
3. **Scalability**: No need to redeploy code to add new items
4. **Maintainability**: Centralized item management through database operations
5. **Admin Tools**: Can build admin interfaces to manage items dynamically

## Testing
- All Go packages compile successfully
- Database migration is properly structured with up/down scripts
- Fallback mechanisms ensure system stability during transition

## Next Steps
1. Run the migration in development/staging environment
2. Test the `/sonar/items` endpoint to verify database integration
3. Verify inventory item operations still work correctly
4. Consider building admin tools for dynamic item management
5. Remove hardcoded items once database migration is fully stable