-- Update inventory_items table structure to match the hardcoded items
-- Create enum for equipment slots

CREATE TYPE rarity_tier_type AS ENUM (
    'Common',
    'Uncommon',
    'Rare',
    'Epic',
    'Legendary',
    'Mythic'
);

CREATE TYPE equipment_slot_type AS ENUM (
    'head',
    'chest', 
    'left_hand',
    'right_hand',
    'feet',
    'gloves',
    'neck',
    'left_ring',
    'right_ring',
    'legs',
    'consumable',
    'passive',
    'equippable'
);

ALTER TABLE inventory_items ADD COLUMN rarity_tier rarity_tier_type NOT NULL DEFAULT 'Common';
ALTER TABLE inventory_items ADD COLUMN equipment_slot equipment_slot_type;
ALTER TABLE inventory_items ADD COLUMN min_damage INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN max_damage INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN defense INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN health INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN speed INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN crit_chance INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN crit_damage INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN attack_range INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN damage_type TEXT DEFAULT 'physical';
ALTER TABLE inventory_items ADD COLUMN plus_strength INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN plus_agility INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN plus_intelligence INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN plus_wisdom INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN plus_constitution INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN plus_charisma INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN permanant_identifier TEXT DEFAULT NULL;

-- Add missing properties for comprehensive item system
ALTER TABLE inventory_items ADD COLUMN weight DECIMAL(5,2) DEFAULT 0.0;
ALTER TABLE inventory_items ADD COLUMN value INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN durability INTEGER DEFAULT 100;
ALTER TABLE inventory_items ADD COLUMN max_durability INTEGER DEFAULT 100;
ALTER TABLE inventory_items ADD COLUMN level_requirement INTEGER DEFAULT 1;
ALTER TABLE inventory_items ADD COLUMN stackable BOOLEAN DEFAULT false;
ALTER TABLE inventory_items ADD COLUMN max_stack_size INTEGER DEFAULT 1;
ALTER TABLE inventory_items ADD COLUMN tradeable BOOLEAN DEFAULT true;
ALTER TABLE inventory_items ADD COLUMN cooldown INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN charges INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN max_charges INTEGER DEFAULT 0;
ALTER TABLE inventory_items ADD COLUMN quest_related BOOLEAN DEFAULT false;
ALTER TABLE inventory_items ADD COLUMN crafting_ingredients JSONB DEFAULT NULL;
ALTER TABLE inventory_items ADD COLUMN special_abilities JSONB DEFAULT NULL;
ALTER TABLE inventory_items ADD COLUMN item_color TEXT DEFAULT NULL;
ALTER TABLE inventory_items ADD COLUMN animation_effects TEXT DEFAULT NULL;
ALTER TABLE inventory_items ADD COLUMN sound_effects TEXT DEFAULT NULL;
ALTER TABLE inventory_items ADD COLUMN bonus_stats JSONB DEFAULT NULL;

-- Insert sample items with enhanced properties
INSERT INTO inventory_items (
    id, created_at, updated_at, name, image_url, flavor_text, effect_text, 
    rarity_tier, equipment_slot, min_damage, max_damage, defense, health, 
    speed, crit_chance, crit_damage, attack_range, damage_type, 
    plus_strength, plus_agility, plus_intelligence, plus_wisdom, 
    plus_constitution, plus_charisma, permanant_identifier, weight, value, 
    durability, max_durability, level_requirement, stackable, max_stack_size, 
    tradeable, cooldown, charges, max_charges, quest_related, 
    crafting_ingredients, special_abilities, item_color, animation_effects, 
    sound_effects, bonus_stats
) VALUES 
(14, NOW(), NOW(), 'Wicked Spellbook', 'https://crew-points-of-interest.s3.us-east-1.amazonaws.com/wicked-spellbook.png', 'The spellbook whispers to you. Ignore it.', 'Steal all of another team''s items.', 'Epic', 'left_hand', 0, 0, 0, 0, 0, 0, 0, 0, 'magical', 0, 0, 5, 0, 0, 0, 'wicked_spellbook_001', 2.5, 1500, 100, 100, 10, false, 1, true, 300, 0, 0, false, NULL, '{"steal_items": true, "whispers": true}', 'purple', 'whispering_pages', 'dark_whispers', '{"intelligence": 5, "magic_power": 10}'),
(15, NOW(), NOW(), 'The Compass of Peace', 'https://crew-points-of-interest.s3.us-east-1.amazonaws.com/compass-of-peace.png', 'Given to you by Shalimar the Merchant. The compass is said to point towards what the wearer needs most to heal.', 'Negate up to 3 damage when held.', 'Rare', 'neck', 0, 0, 3, 0, 0, 0, 0, 0, 'none', 0, 0, 0, 2, 0, 0, 'compass_peace_001', 1.0, 800, 100, 100, 5, false, 1, true, 0, 0, 0, false, NULL, '{"damage_negation": 3, "healing_guidance": true}', 'gold', 'gentle_glow', 'soft_chime', '{"wisdom": 2, "healing": 3}'),
(16, NOW(), NOW(), 'Pirate''s Tricorn Hat', 'https://crew-points-of-interest.s3.amazonaws.com/tricorn-hat.png', 'A weathered hat that has seen many adventures on the high seas. Its feathers still dance in the wind.', 'Increases treasure finding by 10% when worn.', 'Uncommon', 'head', 0, 0, 2, 0, 0, 0, 0, 0, 'none', 0, 1, 0, 0, 0, 0, 'tricorn_hat_001', 0.8, 250, 85, 100, 3, false, 1, true, 0, 0, 0, false, NULL, '{"treasure_finding": 10}', 'brown', 'feathers_dance', 'wind_rustle', '{"agility": 1, "luck": 5}'),
(17, NOW(), NOW(), 'Captain''s Coat', 'https://crew-points-of-interest.s3.amazonaws.com/captains-coat.png', 'A noble coat worn by a captain of renown. Its brass buttons still gleam despite the salt and spray.', 'Provides +5 defense against damage when worn.', 'Epic', 'chest', 0, 0, 5, 0, 0, 0, 0, 0, 'none', 0, 0, 0, 0, 2, 0, 'captains_coat_001', 3.2, 1200, 90, 100, 8, false, 1, true, 0, 0, 0, false, NULL, '{"defense_boost": 5, "captain_aura": true}', 'navy_blue', 'brass_shine', 'fabric_swish', '{"constitution": 2, "defense": 5}');

-- Starter Equipment Items
INSERT INTO inventory_items (
    id, created_at, updated_at, name, image_url, flavor_text, effect_text, 
    rarity_tier, equipment_slot, min_damage, max_damage, defense, health, 
    speed, crit_chance, crit_damage, attack_range, damage_type, 
    plus_strength, plus_agility, plus_intelligence, plus_wisdom, 
    plus_constitution, plus_charisma, permanant_identifier, weight, value, 
    durability, max_durability, level_requirement, stackable, max_stack_size, 
    tradeable, cooldown, charges, max_charges, quest_related, 
    crafting_ingredients, special_abilities, item_color, animation_effects, 
    sound_effects, bonus_stats
) VALUES 
-- Weapons
(NOW(), NOW(), 'Rusty Dagger', 'https://crew-points-of-interest.s3.amazonaws.com/rusty-dagger.png', 'A simple iron dagger with a worn leather grip. Not much to look at, but it gets the job done.', 'Deals 2-4 damage to enemies.', 'Common', 'right_hand', 2, 4, 0, 0, 0, 5, 0, 1, 'physical', 0, 0, 0, 0, 0, 0, 'rusty_dagger_001', 1.2, 50, 75, 100, 1, false, 1, true, 0, 0, 0, false, NULL, '{"quick_strike": true}', 'iron', 'metal_shine', 'blade_swish', '{"agility": 1}'),
(NOW(), NOW(), 'Wooden Staff', 'https://crew-points-of-interest.s3.amazonaws.com/wooden-staff.png', 'A sturdy oak staff, carved with simple runes. Perfect for channeling magical energy.', 'Deals 1-3 damage and +2 intelligence.', 'Common', 'left_hand', 1, 3, 0, 0, 0, 0, 0, 2, 'magical', 0, 0, 2, 0, 0, 0, 'wooden_staff_001', 2.8, 75, 90, 100, 1, false, 1, true, 0, 0, 0, false, NULL, '{"magic_channel": true}', 'brown', 'rune_glow', 'wooden_thud', '{"intelligence": 2, "magic_power": 3}'),
(NOW(), NOW(), 'Training Bow', 'https://crew-points-of-interest.s3.amazonaws.com/training-bow.png', 'A simple yew bow used by novice archers. The string is well-worn but reliable.', 'Deals 3-5 damage with increased range.', 'Common', 'left_hand', 3, 5, 0, 0, 0, 8, 0, 3, 'physical', 0, 2, 0, 0, 0, 0, 'training_bow_001', 1.5, 60, 80, 100, 1, false, 1, true, 0, 0, 0, false, NULL, '{"extended_range": true}', 'brown', 'string_twang', 'arrow_whistle', '{"agility": 2, "range": 3}'),

-- Armor
(NOW(), NOW(), 'Leather Jerkin', 'https://crew-points-of-interest.s3.amazonaws.com/leather-jerkin.png', 'A simple leather vest that provides basic protection without restricting movement.', 'Provides +2 defense and +1 agility.', 'Common', 'chest', 0, 0, 2, 0, 0, 0, 0, 0, 'none', 0, 1, 0, 0, 0, 0, 'leather_jerkin_001', 2.0, 80, 85, 100, 1, false, 1, true, 0, 0, 0, false, NULL, '{"light_armor": true}', 'brown', 'leather_shine', 'soft_rustle', '{"defense": 2, "agility": 1}'),
(NOW(), NOW(), 'Cloth Robes', 'https://crew-points-of-interest.s3.amazonaws.com/cloth-robes.png', 'Simple robes made of sturdy cotton. They offer little protection but are comfortable for spellcasting.', 'Provides +1 defense and +2 intelligence.', 'Common', 'chest', 0, 0, 1, 0, 0, 0, 0, 0, 'none', 0, 0, 2, 0, 0, 0, 'cloth_robes_001', 1.5, 45, 90, 100, 1, false, 1, true, 0, 0, 0, false, NULL, '{"spellcasting_focus": true}', 'white', 'fabric_flow', 'soft_swish', '{"defense": 1, "intelligence": 2}'),
(NOW(), NOW(), 'Iron Gauntlets', 'https://crew-points-of-interest.s3.amazonaws.com/iron-gauntlets.png', 'Heavy iron gauntlets that protect your hands in combat. They''re a bit cumbersome but very protective.', 'Provides +3 defense and +1 strength.', 'Common', 'gloves', 0, 0, 3, 0, 0, 0, 0, 0, 'none', 1, 0, 0, 0, 0, 0, 'iron_gauntlets_001', 2.5, 120, 95, 100, 2, false, 1, true, 0, 0, 0, false, NULL, '{"hand_protection": true}', 'iron', 'metal_clank', 'iron_thud', '{"defense": 3, "strength": 1}'),
(NOW(), NOW(), 'Leather Boots', 'https://crew-points-of-interest.s3.amazonaws.com/leather-boots.png', 'Well-made leather boots that provide good traction and basic foot protection.', 'Provides +1 defense and +2 speed.', 'Common', 'feet', 0, 0, 1, 0, 2, 0, 0, 0, 'none', 0, 0, 0, 0, 0, 0, 'leather_boots_001', 1.8, 65, 90, 100, 1, false, 1, true, 0, 0,
(NOW(), NOW(), 'Copper Ring', 'https://crew-points-of-interest.s3.amazonaws.com/copper-ring.png', 'A simple copper ring with a small gemstone. It provides a minor magical boost.', 'Provides +1 to all stats.', 'Common', 'left_ring', 0, 0, 0, 0, 0, 0, 0, 0, 'none', 1, 1, 1, 1, 1, 1, 'copper_ring_001', 0.1, 100, 100, 100, 1, false, 1, true, 0, 0, 0, false, NULL, '{"minor_boost": true}', 'copper', 'gem_sparkle', 'metal_clink', '{"all_stats": 1}'),

-- Consumable Items
(NOW(), NOW(), 'Health Potion', 'https://crew-points-of-interest.s3.amazonaws.com/health-potion.png', 'A red liquid that bubbles gently. It smells sweet and restorative.', 'Restores 25 health points when consumed.', 'Common', 'consumable', 0, 0, 0, 25, 0, 0, 0, 0, 'none', 0, 0, 0, 0, 0, 0, 'health_potion_001', 0.5, 25, 100, 100, 1, true, 10, true, 0, 1, 1, false, NULL, '{"healing": 25}', 'red', 'bubbling', 'liquid_slosh', '{"healing": 25}'),
(NOW(), NOW(), 'Mana Potion', 'https://crew-points-of-interest.s3.amazonaws.com/mana-potion.png', 'A blue liquid that glows with magical energy. It tingles when you touch it.', 'Restores 30 mana points when consumed.', 'Common', 'consumable', 0, 0, 0, 0, 0, 0, 0, 0, 'none', 0, 0, 0, 0, 0, 0, 'mana_potion_001', 0.5, 30, 100, 100, 1, true, 10, true, 0, 1, 1, false, NULL, '{"mana_restore": 30}', 'blue', 'magical_glow', 'magical_hum', '{"mana": 30}'),
(NOW(), NOW(), 'Bread Loaf', 'https://crew-points-of-interest.s3.amazonaws.com/bread-loaf.png', 'A fresh loaf of bread with a golden crust. It smells warm and comforting.', 'Restores 15 health and provides temporary +1 constitution.', 'Common', 'consumable', 0, 0, 0, 15, 0, 0, 0, 0, 'none', 0, 0, 0, 0, 1, 0, 'bread_loaf_001', 0.8, 5, 100, 100, 1, true, 5, true, 0, 1, 1, false, NULL, '{"nourishment": true}', 'golden', 'steam_rise', 'soft_crunch', '{"health": 15, "constitution": 1}'),
(NOW(), NOW(), 'Antidote', 'https://crew-points-of-interest.s3.amazonaws.com/antidote.png', 'A clear liquid with a bitter taste. It neutralizes most common poisons.', 'Removes poison effects and provides poison resistance for 1 hour.', 'Uncommon', 'consumable', 0, 0, 0, 0, 0, 0, 0, 0, 'none', 0, 0, 0, 0, 0
(NOW(), NOW(), 'Strength Elixir', 'https://crew-points-of-interest.s3.amazonaws.com/strength-elixir.png', 'A thick, red liquid that tastes like iron and fire. It makes your muscles feel like steel.', 'Provides +5 strength for 1 hour.', 'Uncommon', 'consumable', 0, 0, 0, 0, 0, 0, 0, 0, 'none', 5, 0, 0, 0, 0, 0, 'strength_elixir_001', 0.6, 150, 100, 100, 4, true, 3, true, 600, 1, 1, false, NULL, '{"strength_boost": 3600}', 'red', 'muscle_glow', 'powerful_roar', '{"strength": 5}'),
(NOW(), NOW(), 'Speed Draught', 'https://crew-points-of-interest.s3.amazonaws.com/speed-draught.png', 'A fizzy, green liquid that bubbles and pops. It makes you feel incredibly light and fast.', 'Provides +4 speed for 30 minutes.', 'Uncommon', 'consumable', 0, 0, 0, 0, 4, 0, 0, 0, 'none', 0, 0, 0, 0, 0, 0, 'speed_draught_001', 0.4, 120, 100, 100, 3, true, 3, true, 0, 1, 1, false, NULL, '{"speed_boost": 1800}', 'green', 'speed_trails', 'wind_whistle', '{"speed": 4}'),
(NOW(), NOW(), 'Lucky Charm', 'https://crew-points-of-interest.s3.amazonaws.com/lucky-charm.png', 'A small, four-leaf clover preserved in resin. It seems to glow with good fortune.', 'Increases luck and critical hit chance for 1 hour.', 'Rare', 'consumable', 0, 0, 0, 0, 0, 10, 0, 0, 'none', 0, 0, 0, 0, 0, 0, 'lucky_charm_001', 0.2, 300, 100, 100, 6, true, 1, true, 3600, 1, 1, false, NULL, '{"luck_boost": 3600, "crit_boost": 3600}', 'green', 'fortune_glow', 'lucky_chime', '{"luck": 10, "crit_chance": 10}');
