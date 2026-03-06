DROP TABLE IF EXISTS monster_battle_invites;
DROP TABLE IF EXISTS monster_battle_participants;

ALTER TABLE monster_battles
DROP COLUMN IF EXISTS turn_index;

ALTER TABLE monster_battles
DROP COLUMN IF EXISTS state;
