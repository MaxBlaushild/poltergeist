ALTER TABLE monster_battle_participants
ADD COLUMN IF NOT EXISTS base_resources_awarded JSONB NOT NULL DEFAULT '[]'::jsonb;
