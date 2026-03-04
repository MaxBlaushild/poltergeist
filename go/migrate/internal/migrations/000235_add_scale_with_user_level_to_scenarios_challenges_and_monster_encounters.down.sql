ALTER TABLE monster_encounters
  DROP COLUMN IF EXISTS scale_with_user_level;

ALTER TABLE challenges
  DROP COLUMN IF EXISTS scale_with_user_level;

ALTER TABLE scenarios
  DROP COLUMN IF EXISTS scale_with_user_level;
