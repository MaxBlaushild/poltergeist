ALTER TABLE monster_battle_participants
ADD COLUMN IF NOT EXISTS reward_experience integer NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS reward_gold integer NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS items_awarded jsonb NOT NULL DEFAULT '[]'::jsonb;
