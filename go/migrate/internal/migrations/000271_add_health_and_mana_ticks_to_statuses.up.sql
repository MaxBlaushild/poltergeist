ALTER TABLE user_statuses
  ADD COLUMN IF NOT EXISTS health_per_tick integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS mana_per_tick integer NOT NULL DEFAULT 0;

ALTER TABLE monster_statuses
  ADD COLUMN IF NOT EXISTS health_per_tick integer NOT NULL DEFAULT 0;
