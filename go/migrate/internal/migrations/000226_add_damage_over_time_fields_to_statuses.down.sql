ALTER TABLE monster_statuses
DROP COLUMN IF EXISTS last_tick_at,
DROP COLUMN IF EXISTS damage_per_tick;

ALTER TABLE user_statuses
DROP COLUMN IF EXISTS last_tick_at,
DROP COLUMN IF EXISTS damage_per_tick;
