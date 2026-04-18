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

ALTER TABLE inventory_items
  ADD COLUMN genre_id UUID;

ALTER TABLE inventory_item_suggestion_jobs
  ADD COLUMN genre_id UUID;

UPDATE inventory_items
SET genre_id = fantasy.id
FROM (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC
  LIMIT 1
) fantasy
WHERE inventory_items.genre_id IS NULL;

UPDATE inventory_item_suggestion_jobs
SET genre_id = fantasy.id
FROM (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC
  LIMIT 1
) fantasy
WHERE inventory_item_suggestion_jobs.genre_id IS NULL;

ALTER TABLE inventory_items
  ALTER COLUMN genre_id SET NOT NULL,
  ADD CONSTRAINT inventory_items_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES zone_genres(id);

ALTER TABLE inventory_item_suggestion_jobs
  ALTER COLUMN genre_id SET NOT NULL,
  ADD CONSTRAINT inventory_item_suggestion_jobs_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES zone_genres(id);

CREATE INDEX idx_inventory_items_genre_id
  ON inventory_items (genre_id, archived, name);

CREATE INDEX idx_inventory_item_suggestion_jobs_genre_id
  ON inventory_item_suggestion_jobs (genre_id, created_at DESC);

COMMIT;
