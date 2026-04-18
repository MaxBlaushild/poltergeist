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

ALTER TABLE points_of_interest
  ADD COLUMN genre_id UUID;

ALTER TABLE point_of_interest_imports
  ADD COLUMN genre_id UUID;

UPDATE points_of_interest
SET genre_id = fantasy.id
FROM (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC
  LIMIT 1
) fantasy
WHERE points_of_interest.genre_id IS NULL;

UPDATE point_of_interest_imports
SET genre_id = fantasy.id
FROM (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC
  LIMIT 1
) fantasy
WHERE point_of_interest_imports.genre_id IS NULL;

ALTER TABLE points_of_interest
  ALTER COLUMN genre_id SET NOT NULL,
  ADD CONSTRAINT points_of_interest_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES zone_genres(id);

ALTER TABLE point_of_interest_imports
  ALTER COLUMN genre_id SET NOT NULL,
  ADD CONSTRAINT point_of_interest_imports_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES zone_genres(id);

CREATE INDEX idx_points_of_interest_genre_id ON points_of_interest(genre_id);
CREATE INDEX idx_point_of_interest_imports_genre_id ON point_of_interest_imports(genre_id);

COMMIT;
