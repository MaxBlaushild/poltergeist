BEGIN;

INSERT INTO zone_genres (id, created_at, updated_at, name, sort_order, active, prompt_seed)
SELECT
  uuid_generate_v4(),
  NOW(),
  NOW(),
  'Fantasy',
  0,
  TRUE,
  'Keep the genre framing grounded in classic fantasy action RPG adventure: mythic beasts, arcane magic, dungeon ecology, swords-and-sorcery threats, and medieval-adjacent weapons, armor, and factions.'
WHERE NOT EXISTS (
  SELECT 1 FROM zone_genres WHERE LOWER(name) = LOWER('Fantasy')
);

ALTER TABLE characters
  ADD COLUMN genre_id UUID;

ALTER TABLE character_templates
  ADD COLUMN genre_id UUID;

UPDATE characters
SET genre_id = fantasy.id
FROM (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC
  LIMIT 1
) fantasy
WHERE characters.genre_id IS NULL;

UPDATE character_templates
SET genre_id = fantasy.id
FROM (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC
  LIMIT 1
) fantasy
WHERE character_templates.genre_id IS NULL;

ALTER TABLE characters
  ALTER COLUMN genre_id SET NOT NULL,
  ADD CONSTRAINT characters_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES zone_genres(id);

ALTER TABLE character_templates
  ALTER COLUMN genre_id SET NOT NULL,
  ADD CONSTRAINT character_templates_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES zone_genres(id);

CREATE INDEX idx_characters_genre_id
  ON characters (genre_id, name);

CREATE INDEX idx_character_templates_genre_id
  ON character_templates (genre_id, name);

COMMIT;
