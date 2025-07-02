-- Add back the JSONB columns to monsters table
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