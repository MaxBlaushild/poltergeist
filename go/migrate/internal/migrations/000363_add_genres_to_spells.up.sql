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

ALTER TABLE spells
  ADD COLUMN genre_id UUID;

UPDATE spells
SET genre_id = fantasy.id
FROM (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC
  LIMIT 1
) fantasy
WHERE spells.genre_id IS NULL;

ALTER TABLE spells
  ALTER COLUMN genre_id SET NOT NULL,
  ADD CONSTRAINT spells_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES zone_genres(id);

CREATE INDEX idx_spells_genre_id
  ON spells (genre_id, ability_type, name);

COMMIT;
