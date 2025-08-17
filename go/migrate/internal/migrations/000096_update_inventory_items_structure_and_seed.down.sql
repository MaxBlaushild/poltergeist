-- Remove all seeded data
DELETE FROM inventory_items WHERE id IN (14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33);

-- Remove the new columns (in reverse order of creation)
ALTER TABLE inventory_items DROP COLUMN bonus_stats;
ALTER TABLE inventory_items DROP COLUMN sound_effects;
ALTER TABLE inventory_items DROP COLUMN animation_effects;
ALTER TABLE inventory_items DROP COLUMN item_color;
ALTER TABLE inventory_items DROP COLUMN special_abilities;
ALTER TABLE inventory_items DROP COLUMN crafting_ingredients;
ALTER TABLE inventory_items DROP COLUMN quest_related;
ALTER TABLE inventory_items DROP COLUMN max_charges;
ALTER TABLE inventory_items DROP COLUMN charges;
ALTER TABLE inventory_items DROP COLUMN cooldown;
ALTER TABLE inventory_items DROP COLUMN tradeable;
ALTER TABLE inventory_items DROP COLUMN max_stack_size;
ALTER TABLE inventory_items DROP COLUMN stackable;
ALTER TABLE inventory_items DROP COLUMN level_requirement;
ALTER TABLE inventory_items DROP COLUMN max_durability;
ALTER TABLE inventory_items DROP COLUMN durability;
ALTER TABLE inventory_items DROP COLUMN value;
ALTER TABLE inventory_items DROP COLUMN weight;

ALTER TABLE inventory_items DROP COLUMN permanant_identifier;
ALTER TABLE inventory_items DROP COLUMN plus_charisma;
ALTER TABLE inventory_items DROP COLUMN plus_constitution;
ALTER TABLE inventory_items DROP COLUMN plus_wisdom;
ALTER TABLE inventory_items DROP COLUMN plus_intelligence;
ALTER TABLE inventory_items DROP COLUMN plus_agility;
ALTER TABLE inventory_items DROP COLUMN plus_strength;
ALTER TABLE inventory_items DROP COLUMN damage_type;
ALTER TABLE inventory_items DROP COLUMN attack_range;
ALTER TABLE inventory_items DROP COLUMN crit_damage;
ALTER TABLE inventory_items DROP COLUMN crit_chance;
ALTER TABLE inventory_items DROP COLUMN speed;
ALTER TABLE inventory_items DROP COLUMN health;
ALTER TABLE inventory_items DROP COLUMN defense;
ALTER TABLE inventory_items DROP COLUMN max_damage;
ALTER TABLE inventory_items DROP COLUMN min_damage;
ALTER TABLE inventory_items DROP COLUMN equipment_slot;
ALTER TABLE inventory_items DROP COLUMN rarity_tier;

-- Drop the custom types
DROP TYPE equipment_slot_type;
DROP TYPE rarity_tier_type;

-- Change id back to UUID
ALTER TABLE inventory_items DROP CONSTRAINT inventory_items_pkey;
ALTER TABLE inventory_items DROP COLUMN id;
ALTER TABLE inventory_items ADD COLUMN id UUID PRIMARY KEY;