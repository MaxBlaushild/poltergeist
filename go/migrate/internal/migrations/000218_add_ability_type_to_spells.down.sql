DROP INDEX IF EXISTS idx_spells_ability_type;

ALTER TABLE spells
DROP COLUMN IF EXISTS ability_type;
