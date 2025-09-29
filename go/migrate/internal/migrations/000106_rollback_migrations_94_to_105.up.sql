-- Rollback migration 104: Convert inventory_item_id in point_of_interest_challenges back to integer
ALTER TABLE point_of_interest_challenges DROP CONSTRAINT IF EXISTS fk_point_of_interest_challenges_inventory_item_id;
ALTER TABLE point_of_interest_challenges DROP COLUMN IF EXISTS inventory_item_id;
ALTER TABLE point_of_interest_challenges ADD COLUMN inventory_item_id INTEGER;

-- Rollback migration 103: Convert reward in quest_archetype_challenges back to integer
ALTER TABLE quest_archetype_challenges DROP CONSTRAINT IF EXISTS fk_quest_archetype_challenges_reward;
ALTER TABLE quest_archetype_challenges DROP COLUMN IF EXISTS reward;
ALTER TABLE quest_archetype_challenges ADD COLUMN reward INTEGER;

-- Rollback migration 102: Convert inventory_item_id in owned_inventory_items back to integer
ALTER TABLE owned_inventory_items DROP CONSTRAINT IF EXISTS fk_owned_inventory_items_inventory_item_id;
ALTER TABLE owned_inventory_items DROP COLUMN IF EXISTS inventory_item_id;
ALTER TABLE owned_inventory_items ADD COLUMN inventory_item_id INTEGER;

-- Rollback migration 101: Drop monster_actions table and restore JSONB columns to monsters
-- First, restore the JSONB columns to monsters table
ALTER TABLE monsters ADD COLUMN IF NOT EXISTS special_abilities JSONB;
ALTER TABLE monsters ADD COLUMN IF NOT EXISTS actions JSONB;
ALTER TABLE monsters ADD COLUMN IF NOT EXISTS legendary_actions JSONB;
ALTER TABLE monsters ADD COLUMN IF NOT EXISTS reactions JSONB;

-- Migrate data back from monster_actions to monsters table JSONB columns
DO $$
DECLARE
    monster_record RECORD;
    actions_array JSONB := '[]';
    special_abilities_array JSONB := '[]';
    legendary_actions_array JSONB := '[]';
    reactions_array JSONB := '[]';
    action_record RECORD;
    action_json JSONB;
BEGIN
    -- Loop through all monsters
    FOR monster_record IN SELECT DISTINCT monster_id FROM monster_actions LOOP
        
        -- Reset arrays
        actions_array := '[]';
        special_abilities_array := '[]';
        legendary_actions_array := '[]';
        reactions_array := '[]';
        
        -- Collect actions
        FOR action_record IN 
            SELECT * FROM monster_actions 
            WHERE monster_id = monster_record.monster_id AND action_type = 'action' AND active = true
            ORDER BY order_index
        LOOP
            action_json := jsonb_build_object(
                'name', action_record.name,
                'description', action_record.description
            );
            
            IF action_record.attack_bonus IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('attack_bonus', action_record.attack_bonus);
            END IF;
            
            IF action_record.damage_dice IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('damage', action_record.damage_dice);
            END IF;
            
            IF action_record.damage_type IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('damage_type', action_record.damage_type);
            END IF;
            
            IF action_record.additional_damage_dice IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('additional_damage', action_record.additional_damage_dice);
            END IF;
            
            IF action_record.save_dc IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('save_dc', action_record.save_dc);
            END IF;
            
            IF action_record.save_ability IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('save_ability', action_record.save_ability);
            END IF;
            
            IF action_record.special_effects IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('special', action_record.special_effects);
            END IF;
            
            IF action_record.recharge IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('recharge', action_record.recharge);
            END IF;
            
            actions_array := actions_array || action_json;
        END LOOP;
        
        -- Collect special abilities
        FOR action_record IN 
            SELECT * FROM monster_actions 
            WHERE monster_id = monster_record.monster_id AND action_type = 'special_ability' AND active = true
            ORDER BY order_index
        LOOP
            action_json := jsonb_build_object(
                'name', action_record.name,
                'description', action_record.description
            );
            
            IF action_record.special_effects IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('special', action_record.special_effects);
            END IF;
            
            IF action_record.recharge IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('recharge', action_record.recharge);
            END IF;
            
            special_abilities_array := special_abilities_array || action_json;
        END LOOP;
        
        -- Collect legendary actions
        FOR action_record IN 
            SELECT * FROM monster_actions 
            WHERE monster_id = monster_record.monster_id AND action_type = 'legendary_action' AND active = true
            ORDER BY order_index
        LOOP
            action_json := jsonb_build_object(
                'name', action_record.name,
                'description', action_record.description
            );
            
            IF action_record.attack_bonus IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('attack_bonus', action_record.attack_bonus);
            END IF;
            
            IF action_record.damage_dice IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('damage', action_record.damage_dice);
            END IF;
            
            IF action_record.damage_type IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('damage_type', action_record.damage_type);
            END IF;
            
            IF action_record.special_effects IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('special', action_record.special_effects);
            END IF;
            
            IF action_record.legendary_cost IS NOT NULL AND action_record.legendary_cost != 1 THEN
                action_json := action_json || jsonb_build_object('legendary_cost', action_record.legendary_cost);
            END IF;
            
            legendary_actions_array := legendary_actions_array || action_json;
        END LOOP;
        
        -- Collect reactions
        FOR action_record IN 
            SELECT * FROM monster_actions 
            WHERE monster_id = monster_record.monster_id AND action_type = 'reaction' AND active = true
            ORDER BY order_index
        LOOP
            action_json := jsonb_build_object(
                'name', action_record.name,
                'description', action_record.description
            );
            
            IF action_record.special_effects IS NOT NULL THEN
                action_json := action_json || jsonb_build_object('special', action_record.special_effects);
            END IF;
            
            reactions_array := reactions_array || action_json;
        END LOOP;
        
        -- Update the monster with the collected arrays
        UPDATE monsters SET
            actions = CASE WHEN jsonb_array_length(actions_array) > 0 THEN actions_array ELSE NULL END,
            special_abilities = CASE WHEN jsonb_array_length(special_abilities_array) > 0 THEN special_abilities_array ELSE NULL END,
            legendary_actions = CASE WHEN jsonb_array_length(legendary_actions_array) > 0 THEN legendary_actions_array ELSE NULL END,
            reactions = CASE WHEN jsonb_array_length(reactions_array) > 0 THEN reactions_array ELSE NULL END
        WHERE id = monster_record.monster_id;
        
    END LOOP;
END $$;

-- Drop the monster_actions table
DROP INDEX IF EXISTS idx_monster_actions_order;
DROP INDEX IF EXISTS idx_monster_actions_active;
DROP INDEX IF EXISTS idx_monster_actions_type;
DROP INDEX IF EXISTS idx_monster_actions_monster_id;
DROP TABLE IF EXISTS monster_actions;

-- Rollback migration 100: Drop monsters table
DROP INDEX IF EXISTS idx_monsters_active;
DROP INDEX IF EXISTS idx_monsters_size;
DROP INDEX IF EXISTS idx_monsters_type;
DROP INDEX IF EXISTS idx_monsters_challenge_rating;
DROP TABLE IF EXISTS monsters;

-- Rollback migration 98: Remove dnd_class_id from users table
DROP INDEX IF EXISTS idx_users_dnd_class_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_dnd_class;
ALTER TABLE users DROP COLUMN IF EXISTS dnd_class_id;

-- Rollback migration 97: Drop dnd_classes table
DROP TABLE IF EXISTS dnd_classes;

-- Rollback migration 96: Remove inventory items structure changes and seeded data
-- Remove all seeded data
DELETE FROM inventory_items WHERE permanant_identifier IN (
    'wicked_spellbook_001', 'compass_peace_001', 'tricorn_hat_001', 'captains_coat_001',
    'rusty_dagger_001', 'wooden_staff_001', 'training_bow_001', 'leather_jerkin_001',
    'cloth_robes_001', 'iron_gauntlets_001', 'leather_boots_001', 'copper_ring_001',
    'health_potion_001', 'mana_potion_001', 'bread_loaf_001', 'antidote_001',
    'invisibility_potion_001', 'strength_elixir_001', 'speed_draught_001', 'lucky_charm_001'
);

-- Remove the new columns (in reverse order of creation)
ALTER TABLE inventory_items DROP COLUMN IF EXISTS bonus_stats;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS sound_effects;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS animation_effects;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS item_color;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS special_abilities;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS crafting_ingredients;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS quest_related;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS max_charges;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS charges;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS cooldown;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS tradeable;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS max_stack_size;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS stackable;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS level_requirement;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS max_durability;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS durability;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS value;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS weight;

ALTER TABLE inventory_items DROP COLUMN IF EXISTS permanant_identifier;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS plus_charisma;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS plus_constitution;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS plus_wisdom;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS plus_intelligence;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS plus_agility;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS plus_strength;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS damage_type;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS attack_range;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS crit_damage;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS crit_chance;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS speed;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS health;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS defense;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS max_damage;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS min_damage;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS equipment_slot;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS rarity_tier;

-- Drop the custom types
DROP TYPE IF EXISTS equipment_slot_type;
DROP TYPE IF EXISTS rarity_tier_type;

-- Rollback migration 95: Drop user_stats table
DROP TABLE IF EXISTS user_stats;

-- Rollback migration 94: Drop user_equipment table
DROP INDEX IF EXISTS user_equipment_user_id_idx;
DROP TABLE IF EXISTS user_equipment; 