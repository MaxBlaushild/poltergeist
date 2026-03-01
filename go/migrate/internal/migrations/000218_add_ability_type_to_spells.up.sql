ALTER TABLE spells
ADD COLUMN ability_type TEXT NOT NULL DEFAULT 'spell';

UPDATE spells
SET ability_type = 'spell'
WHERE ability_type IS NULL OR TRIM(ability_type) = '';

CREATE INDEX idx_spells_ability_type ON spells(ability_type);
