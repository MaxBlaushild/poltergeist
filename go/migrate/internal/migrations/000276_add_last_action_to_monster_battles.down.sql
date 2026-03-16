ALTER TABLE monster_battles
DROP COLUMN IF EXISTS last_action,
DROP COLUMN IF EXISTS last_action_sequence;
