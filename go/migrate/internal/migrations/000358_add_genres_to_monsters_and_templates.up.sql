ALTER TABLE zone_genres
  ADD COLUMN IF NOT EXISTS prompt_seed TEXT NOT NULL DEFAULT '';

WITH existing_fantasy AS (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC, id ASC
  LIMIT 1
), inserted_fantasy AS (
  INSERT INTO zone_genres (
    id,
    created_at,
    updated_at,
    name,
    sort_order,
    active,
    prompt_seed
  )
  SELECT
    uuid_generate_v4(),
    NOW(),
    NOW(),
    'Fantasy',
    0,
    TRUE,
    'Keep the genre framing grounded in classic fantasy action RPG adventure: mythic beasts, arcane magic, dungeon ecology, swords-and-sorcery threats, and medieval-adjacent weapons, armor, and factions.'
  WHERE NOT EXISTS (SELECT 1 FROM existing_fantasy)
  RETURNING id
), resolved_fantasy AS (
  SELECT id FROM existing_fantasy
  UNION ALL
  SELECT id FROM inserted_fantasy
)
UPDATE zone_genres
SET
  prompt_seed = 'Keep the genre framing grounded in classic fantasy action RPG adventure: mythic beasts, arcane magic, dungeon ecology, swords-and-sorcery threats, and medieval-adjacent weapons, armor, and factions.',
  updated_at = NOW()
WHERE id IN (SELECT id FROM resolved_fantasy)
  AND COALESCE(BTRIM(prompt_seed), '') = '';

ALTER TABLE monster_templates
  ADD COLUMN IF NOT EXISTS genre_id UUID REFERENCES zone_genres(id);

ALTER TABLE monsters
  ADD COLUMN IF NOT EXISTS genre_id UUID REFERENCES zone_genres(id);

WITH fantasy AS (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC, id ASC
  LIMIT 1
)
UPDATE monster_templates
SET genre_id = (SELECT id FROM fantasy)
WHERE genre_id IS NULL;

WITH fantasy AS (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC, id ASC
  LIMIT 1
)
UPDATE monsters AS m
SET genre_id = COALESCE(mt.genre_id, (SELECT id FROM fantasy))
FROM monster_templates AS mt
WHERE m.genre_id IS NULL
  AND m.template_id = mt.id;

WITH fantasy AS (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC, id ASC
  LIMIT 1
)
UPDATE monsters
SET genre_id = (SELECT id FROM fantasy)
WHERE genre_id IS NULL;

ALTER TABLE monster_templates
  ALTER COLUMN genre_id SET NOT NULL;

ALTER TABLE monsters
  ALTER COLUMN genre_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_monster_templates_genre_id
  ON monster_templates (genre_id, monster_type, archived, name);

CREATE INDEX IF NOT EXISTS idx_monsters_genre_id
  ON monsters (genre_id, zone_id, name);
