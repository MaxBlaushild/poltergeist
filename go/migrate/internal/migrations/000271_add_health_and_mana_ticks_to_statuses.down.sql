ALTER TABLE monster_statuses
  DROP COLUMN IF EXISTS health_per_tick;

ALTER TABLE user_statuses
  DROP COLUMN IF EXISTS mana_per_tick,
  DROP COLUMN IF EXISTS health_per_tick;
