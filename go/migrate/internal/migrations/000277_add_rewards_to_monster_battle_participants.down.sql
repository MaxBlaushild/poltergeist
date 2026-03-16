ALTER TABLE monster_battle_participants
DROP COLUMN IF EXISTS items_awarded,
DROP COLUMN IF EXISTS reward_gold,
DROP COLUMN IF EXISTS reward_experience;
