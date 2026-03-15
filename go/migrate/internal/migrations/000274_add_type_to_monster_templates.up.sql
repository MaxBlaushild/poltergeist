ALTER TABLE monster_templates
ADD COLUMN monster_type TEXT NOT NULL DEFAULT 'monster';

UPDATE monster_templates
SET monster_type = 'monster'
WHERE TRIM(COALESCE(monster_type, '')) = '';
