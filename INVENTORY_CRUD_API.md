# Inventory Item CRUD API

This document describes the complete CRUD API for managing inventory items with stats support for the equipment system.

## Overview

The inventory item system supports creating items with stat bonuses that can be applied to users when equipped. Items can have different types (passive, consumable, equippable) and can be assigned to specific equipment slots.

## Equipment Slots

The following equipment slots are supported:
- `head` - Hats, helmets, circlets
- `chest` - Armor, coats, shirts  
- `legs` - Pants, leggings, skirts
- `feet` - Boots, shoes, sandals
- `left_hand` - Shields, off-hand weapons, books
- `right_hand` - Primary weapons, tools
- `neck` - Necklaces, amulets, compasses
- `ring` - Rings
- `belt` - Belts, sashes
- `gloves` - Gloves, gauntlets

## Stats System

Items can provide bonuses to the following D&D-style stats:
- `strength` - Physical power and melee damage
- `dexterity` - Agility, speed, and ranged accuracy
- `constitution` - Health and endurance
- `intelligence` - Magical power and problem solving
- `wisdom` - Perception and insight
- `charisma` - Social skills and leadership

## API Endpoints

### Get All Items (Public)

```
GET /sonar/items
```

Returns all inventory items without stats information. This is the public endpoint that doesn't require authentication.

**Response:**
```json
[
  {
    "id": 1,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z",
    "name": "Steel Sword",
    "imageUrl": "https://example.com/steel-sword.png",
    "flavorText": "A well-forged blade of tempered steel.",
    "effectText": "Increases combat effectiveness.",
    "rarityTier": "Common",
    "isCaptureType": false,
    "itemType": "equippable",
    "equipmentSlot": "right_hand"
  }
]
```

### Get All Items with Stats (Admin)

```
GET /sonar/admin/items
Authorization: Bearer <token>
```

Returns all inventory items including their stat bonuses. Requires authentication.

**Response:**
```json
[
  {
    "id": 1,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z",
    "name": "Steel Sword",
    "imageUrl": "https://example.com/steel-sword.png",
    "flavorText": "A well-forged blade of tempered steel.",
    "effectText": "Increases combat effectiveness.",
    "rarityTier": "Common",
    "isCaptureType": false,
    "itemType": "equippable",
    "equipmentSlot": "right_hand",
    "stats": {
      "id": "uuid",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z",
      "inventoryItemId": 1,
      "strengthBonus": 2,
      "dexterityBonus": 0,
      "constitutionBonus": 0,
      "intelligenceBonus": 0,
      "wisdomBonus": 0,
      "charismaBonus": 0
    }
  }
]
```

### Get Item by ID with Stats (Admin)

```
GET /sonar/admin/items/:id
Authorization: Bearer <token>
```

Returns a specific inventory item by ID including its stat bonuses.

**Parameters:**
- `id` (integer) - The inventory item ID

**Response:**
```json
{
  "id": 1,
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z",
  "name": "Steel Sword",
  "imageUrl": "https://example.com/steel-sword.png",
  "flavorText": "A well-forged blade of tempered steel.",
  "effectText": "Increases combat effectiveness.",
  "rarityTier": "Common",
  "isCaptureType": false,
  "itemType": "equippable",
  "equipmentSlot": "right_hand",
  "stats": {
    "id": "uuid",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z",
    "inventoryItemId": 1,
    "strengthBonus": 2,
    "dexterityBonus": 0,
    "constitutionBonus": 0,
    "intelligenceBonus": 0,
    "wisdomBonus": 0,
    "charismaBonus": 0
  }
}
```

### Create New Item (Admin)

```
POST /sonar/admin/items
Authorization: Bearer <token>
Content-Type: application/json
```

Creates a new inventory item with optional stat bonuses.

**Request Body:**
```json
{
  "name": "Enchanted Leather Gloves",
  "imageUrl": "https://example.com/enchanted-gloves.png",
  "flavorText": "These gloves shimmer with magical energy.",
  "effectText": "Enhances dexterity and provides magical protection.",
  "rarityTier": "Epic",
  "isCaptureType": false,
  "itemType": "equippable",
  "equipmentSlot": "gloves",
  "stats": {
    "strengthBonus": 0,
    "dexterityBonus": 3,
    "constitutionBonus": 1,
    "intelligenceBonus": 0,
    "wisdomBonus": 0,
    "charismaBonus": 0
  }
}
```

**Validation Rules:**
- `name` (required) - Item name
- `imageUrl` (required) - URL to item image
- `flavorText` (optional) - Descriptive text about the item
- `effectText` (optional) - Description of item effects
- `rarityTier` (required) - Must be one of: "Common", "Uncommon", "Epic", "Mythic", "Not Droppable"
- `isCaptureType` (optional) - Boolean indicating if item can be used for instant captures
- `itemType` (required) - Must be one of: "passive", "consumable", "equippable"
- `equipmentSlot` (conditional) - Required for "equippable" items, must be valid slot name
- `stats` (optional) - Stat bonuses object with integer values

**Response:**
```json
{
  "id": 21,
  "createdAt": "2024-01-01T12:00:00Z",
  "updatedAt": "2024-01-01T12:00:00Z",
  "name": "Enchanted Leather Gloves",
  "imageUrl": "https://example.com/enchanted-gloves.png",
  "flavorText": "These gloves shimmer with magical energy.",
  "effectText": "Enhances dexterity and provides magical protection.",
  "rarityTier": "Epic",
  "isCaptureType": false,
  "itemType": "equippable",
  "equipmentSlot": "gloves",
  "stats": {
    "id": "uuid",
    "createdAt": "2024-01-01T12:00:00Z",
    "updatedAt": "2024-01-01T12:00:00Z",
    "inventoryItemId": 21,
    "strengthBonus": 0,
    "dexterityBonus": 3,
    "constitutionBonus": 1,
    "intelligenceBonus": 0,
    "wisdomBonus": 0,
    "charismaBonus": 0
  }
}
```

### Update Item (Admin)

```
PUT /sonar/admin/items/:id
Authorization: Bearer <token>
Content-Type: application/json
```

Updates an existing inventory item and its stat bonuses.

**Parameters:**
- `id` (integer) - The inventory item ID to update

**Request Body:** Same format as Create Item

**Response:** Same format as Create Item

**Notes:**
- If `stats` is provided, it will update or create the item's stats
- If `stats` is omitted or `null`, existing stats will be removed
- All item fields must be provided (full replacement update)

### Delete Item (Admin)

```
DELETE /sonar/admin/items/:id
Authorization: Bearer <token>
```

Deletes an inventory item and its associated stats.

**Parameters:**
- `id` (integer) - The inventory item ID to delete

**Response:**
```json
{
  "message": "item deleted successfully"
}
```

**Notes:**
- Item stats are automatically deleted due to foreign key CASCADE
- Cannot delete items that are currently owned by users (foreign key constraint)

## Example Usage Scenarios

### Creating a Stat-Based Weapon

```bash
curl -X POST "http://localhost:8042/sonar/admin/items" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dragon Slayer Sword",
    "imageUrl": "https://example.com/dragon-slayer.png",
    "flavorText": "Forged from dragon scales and blessed by ancient magic.",
    "effectText": "Massive strength bonus and fire resistance.",
    "rarityTier": "Mythic",
    "isCaptureType": false,
    "itemType": "equippable",
    "equipmentSlot": "right_hand",
    "stats": {
      "strengthBonus": 5,
      "dexterityBonus": 2,
      "constitutionBonus": 3,
      "intelligenceBonus": 0,
      "wisdomBonus": 1,
      "charismaBonus": 2
    }
  }'
```

### Creating a Consumable Item (No Stats)

```bash
curl -X POST "http://localhost:8042/sonar/admin/items" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Health Potion",
    "imageUrl": "https://example.com/health-potion.png",
    "flavorText": "A crimson liquid that glows with healing energy.",
    "effectText": "Restores health when consumed.",
    "rarityTier": "Common",
    "isCaptureType": false,
    "itemType": "consumable"
  }'
```

### Creating a Passive Item with Stats

```bash
curl -X POST "http://localhost:8042/sonar/admin/items" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tome of Wisdom",
    "imageUrl": "https://example.com/wisdom-tome.png",
    "flavorText": "An ancient book containing profound knowledge.",
    "effectText": "Increases wisdom while carried in inventory.",
    "rarityTier": "Uncommon",
    "isCaptureType": false,
    "itemType": "passive",
    "stats": {
      "strengthBonus": 0,
      "dexterityBonus": 0,
      "constitutionBonus": 0,
      "intelligenceBonus": 2,
      "wisdomBonus": 4,
      "charismaBonus": 0
    }
  }'
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "invalid rarity tier. Must be one of: Common, Uncommon, Epic, Mythic, Not Droppable"
}
```

### 401 Unauthorized
```json
{
  "error": "no authenticated user found"
}
```

### 404 Not Found
```json
{
  "error": "item not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "failed to create item stats: database connection error"
}
```

## Integration with Equipment System

Created items can be immediately used with the existing equipment system:

1. **Item Creation** → Items are created in the database with stats
2. **Item Ownership** → Users can receive items via existing `/sonar/users/giveItem` endpoint
3. **Item Equipping** → Users can equip items via existing `/sonar/equipment/equip` endpoint
4. **Stat Application** → Equipped items provide their stat bonuses to users

## Database Schema

### inventory_items
- Basic item information (name, image, rarity, type, slot)
- Links to `inventory_item_stats` via foreign key

### inventory_item_stats  
- Stat bonuses for each D&D stat
- One-to-one relationship with inventory items
- Automatically deleted when item is deleted (CASCADE)

## Migration

To enable this functionality, run migration 000099:

```bash
migrate -path ./go/migrate/internal/migrations -database "postgres://..." up
```

This creates the `inventory_item_stats` table with proper foreign key relationships and indexes.