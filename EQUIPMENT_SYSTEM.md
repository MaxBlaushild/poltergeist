# Equipment System for Sonar

## Overview

The equipment system adds RPG-style equipment functionality to Sonar, allowing users to equip inventory items that provide passive benefits. This system categorizes items into three types: **passive**, **consumable**, and **equippable**.

## Item Types

### Passive Items
- Items that provide benefits while held in inventory
- Examples: Gold Coin (+1 score), Damage (-2 score), Entseed (+3 score, neutralizes damage)
- No equipping required - effects are automatic

### Consumable Items
- Items that are used once and consumed
- Examples: Cipher of the Laughing Monkey (warps clue texts), Ale (removes damage), Flawed Ruby (instant capture)
- Disappear from inventory after use

### Equippable Items
- Items that must be equipped to specific slots to provide benefits
- Examples: Cortez's Cutlass (weapon), Pirate's Tricorn Hat (head), Captain's Coat (chest)
- Must be actively equipped by the user

## Equipment Slots

The system supports typical RPG equipment slots:
- **Head**: Hats, helmets, circlets
- **Chest**: Armor, coats, shirts
- **Legs**: Pants, leggings, skirts
- **Feet**: Boots, shoes, sandals
- **Left Hand**: Shields, off-hand weapons, books
- **Right Hand**: Primary weapons, tools
- **Neck**: Necklaces, amulets, compasses
- **Ring**: Rings (supports multiple if needed)
- **Belt**: Belts, sashes
- **Gloves**: Gloves, gauntlets

## Database Schema

### User Equipment Table
```sql
CREATE TABLE user_equipment (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    equipment_slot VARCHAR(50) NOT NULL,
    owned_inventory_item_id UUID NOT NULL REFERENCES owned_inventory_items(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### Key Constraints
- Each user can only have one item equipped per slot
- Equipped items must be owned by the user
- Items are automatically unequipped when consumed

## API Endpoints

### Get User Equipment
```
GET /sonar/equipment
```
Returns all currently equipped items for the authenticated user.

**Response:**
```json
[
  {
    "id": "uuid",
    "userId": "uuid", 
    "equipmentSlot": "head",
    "ownedInventoryItemId": "uuid",
    "ownedInventoryItem": {
      "id": "uuid",
      "inventoryItemId": 16,
      "quantity": 1
    },
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
]
```

### Equip Item
```
POST /sonar/equipment/equip
```
Equips an owned inventory item to the appropriate slot.

**Request Body:**
```json
{
  "ownedInventoryItemId": "uuid"
}
```

**Response:**
```json
{
  "message": "item equipped successfully"
}
```

### Unequip Item
```
DELETE /sonar/equipment/unequip/:slot
```
Removes the item from the specified equipment slot.

**Response:**
```json
{
  "message": "item unequipped successfully"
}
```

## Example Items

### Equippable Items Added

1. **Pirate's Tricorn Hat** (Head)
   - Increases treasure finding by 10%
   - Rarity: Uncommon

2. **Captain's Coat** (Chest)
   - Provides +5 defense against damage
   - Rarity: Epic

3. **Seafarer's Boots** (Feet)
   - Increases movement speed by 15%
   - Rarity: Common

4. **Enchanted Ring of Fortune** (Ring)
   - Doubles reward chances for treasure hunting
   - Rarity: Mythic

5. **Leather Sailing Gloves** (Gloves)
   - Reduces chance of dropping items by 50%
   - Rarity: Common

### Updated Existing Items

- **Cortez's Cutlass**: Now equippable to right hand
- **Dagger**: Now equippable to left hand
- **Wicked Spellbook**: Now equippable to left hand
- **The Compass of Peace**: Now equippable to neck

## Business Logic

### Equipping Items
1. User must own the item
2. Item must be of type "equippable"
3. If another item is equipped in the same slot, it's automatically unequipped
4. Item is marked as equipped in the user_equipment table

### Using/Consuming Items
1. If an equipped item is consumed, it's automatically unequipped
2. Item quantity is decremented
3. Item effects no longer apply

### Item Effects
- **Passive items**: Effects applied while in inventory
- **Equippable items**: Effects applied only when equipped
- **Consumable items**: One-time effects when used

## Migration

Run migration `000094_create_user_equipment` to add the equipment table:
```bash
migrate -path ./go/migrate/internal/migrations -database "postgres://..." up
```

## Future Enhancements

1. **Multiple Ring Slots**: Support for multiple rings
2. **Equipment Sets**: Bonus effects for wearing complete sets
3. **Item Durability**: Equipment degradation over time
4. **Enchantments**: Temporary or permanent item improvements
5. **Visual Representation**: Equipment visible on user avatars
6. **Equipment Stats**: Detailed stat tracking and comparisons