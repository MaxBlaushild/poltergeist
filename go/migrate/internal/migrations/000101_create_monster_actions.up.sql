CREATE TABLE monster_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    monster_id UUID NOT NULL REFERENCES monsters(id) ON DELETE CASCADE,
    
    -- Action classification
    action_type VARCHAR(50) NOT NULL, -- 'action', 'special_ability', 'legendary_action', 'reaction'
    order_index INTEGER NOT NULL DEFAULT 0, -- For ordering actions within each type
    
    -- Basic action info
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    
    -- Attack mechanics
    attack_bonus INTEGER, -- +4 to hit
    damage_dice VARCHAR(50), -- "1d6+2"
    damage_type VARCHAR(50), -- "slashing", "fire", etc.
    additional_damage_dice VARCHAR(50), -- "4d6"
    additional_damage_type VARCHAR(50), -- "fire"
    
    -- Save mechanics
    save_dc INTEGER, -- DC 15
    save_ability VARCHAR(20), -- "Dexterity"
    save_effect_half_damage BOOLEAN DEFAULT FALSE, -- true if half damage on save
    
    -- Range and area
    range_reach INTEGER, -- 5 ft reach, 80/320 ft range
    range_long INTEGER, -- long range for ranged attacks
    area_type VARCHAR(50), -- "cone", "sphere", "line", "cube"
    area_size INTEGER, -- 60 (for 60-foot cone)
    
    -- Special mechanics
    recharge VARCHAR(10), -- "5-6", "short rest", "long rest"
    uses_per_day INTEGER, -- 3/Day
    special_effects TEXT, -- Additional effects like "target is knocked prone"
    
    -- Legendary action cost
    legendary_cost INTEGER DEFAULT 1, -- How many legendary actions this costs
    
    active BOOLEAN DEFAULT TRUE
);

-- Create indexes for efficient queries
CREATE INDEX idx_monster_actions_monster_id ON monster_actions(monster_id);
CREATE INDEX idx_monster_actions_type ON monster_actions(action_type);
CREATE INDEX idx_monster_actions_active ON monster_actions(active);
CREATE INDEX idx_monster_actions_order ON monster_actions(monster_id, action_type, order_index);

-- Migrate existing action data from monsters table to monster_actions table
DO $$
DECLARE
    monster_record RECORD;
    action_record JSONB;
    action_type TEXT;
    order_idx INTEGER;
BEGIN
    -- Loop through all monsters
    FOR monster_record IN SELECT id, actions, special_abilities, legendary_actions, reactions FROM monsters LOOP
        
        -- Process actions
        IF monster_record.actions IS NOT NULL THEN
            order_idx := 0;
            FOR action_record IN SELECT * FROM jsonb_array_elements(monster_record.actions) LOOP
                INSERT INTO monster_actions (
                    monster_id, action_type, order_index, name, description,
                    attack_bonus, damage_dice, damage_type, additional_damage_dice, additional_damage_type,
                    save_dc, save_ability, special_effects, recharge
                ) VALUES (
                    monster_record.id, 'action', order_idx,
                    action_record->>'name',
                    action_record->>'description',
                    CASE WHEN action_record->>'attack_bonus' != '' THEN (action_record->>'attack_bonus')::INTEGER ELSE NULL END,
                    action_record->>'damage',
                    action_record->>'damage_type',
                    action_record->>'additional_damage',
                    CASE WHEN action_record->>'additional_damage' LIKE '%fire%' THEN 'fire'
                         WHEN action_record->>'additional_damage' LIKE '%cold%' THEN 'cold'
                         WHEN action_record->>'additional_damage' LIKE '%lightning%' THEN 'lightning'
                         WHEN action_record->>'additional_damage' LIKE '%acid%' THEN 'acid'
                         WHEN action_record->>'additional_damage' LIKE '%poison%' THEN 'poison'
                         ELSE NULL END,
                    CASE WHEN action_record->>'save_dc' != '' THEN (action_record->>'save_dc')::INTEGER ELSE NULL END,
                    action_record->>'save_ability',
                    action_record->>'special',
                    action_record->>'recharge'
                );
                order_idx := order_idx + 1;
            END LOOP;
        END IF;
        
        -- Process special abilities
        IF monster_record.special_abilities IS NOT NULL THEN
            order_idx := 0;
            FOR action_record IN SELECT * FROM jsonb_array_elements(monster_record.special_abilities) LOOP
                INSERT INTO monster_actions (
                    monster_id, action_type, order_index, name, description, special_effects, recharge
                ) VALUES (
                    monster_record.id, 'special_ability', order_idx,
                    action_record->>'name',
                    action_record->>'description',
                    action_record->>'special',
                    action_record->>'recharge'
                );
                order_idx := order_idx + 1;
            END LOOP;
        END IF;
        
        -- Process legendary actions
        IF monster_record.legendary_actions IS NOT NULL THEN
            order_idx := 0;
            FOR action_record IN SELECT * FROM jsonb_array_elements(monster_record.legendary_actions) LOOP
                INSERT INTO monster_actions (
                    monster_id, action_type, order_index, name, description,
                    attack_bonus, damage_dice, damage_type, special_effects, legendary_cost
                ) VALUES (
                    monster_record.id, 'legendary_action', order_idx,
                    action_record->>'name',
                    action_record->>'description',
                    CASE WHEN action_record->>'attack_bonus' != '' THEN (action_record->>'attack_bonus')::INTEGER ELSE NULL END,
                    action_record->>'damage',
                    action_record->>'damage_type',
                    action_record->>'special',
                    COALESCE((action_record->>'legendary_cost')::INTEGER, 1)
                );
                order_idx := order_idx + 1;
            END LOOP;
        END IF;
        
        -- Process reactions
        IF monster_record.reactions IS NOT NULL THEN
            order_idx := 0;
            FOR action_record IN SELECT * FROM jsonb_array_elements(monster_record.reactions) LOOP
                INSERT INTO monster_actions (
                    monster_id, action_type, order_index, name, description, special_effects
                ) VALUES (
                    monster_record.id, 'reaction', order_idx,
                    action_record->>'name',
                    action_record->>'description',
                    action_record->>'special'
                );
                order_idx := order_idx + 1;
            END LOOP;
        END IF;
        
    END LOOP;
END $$;

-- Remove the JSONB columns from monsters table since we now have normalized data
ALTER TABLE monsters DROP COLUMN IF EXISTS special_abilities;
ALTER TABLE monsters DROP COLUMN IF EXISTS actions;
ALTER TABLE monsters DROP COLUMN IF EXISTS legendary_actions;
ALTER TABLE monsters DROP COLUMN IF EXISTS reactions;