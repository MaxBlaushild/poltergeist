ALTER TABLE monster_battles
ADD COLUMN last_action_sequence INTEGER NOT NULL DEFAULT 0,
ADD COLUMN last_action JSONB NOT NULL DEFAULT '{}'::jsonb;
