DROP INDEX IF EXISTS idx_monsters_genre_id;
DROP INDEX IF EXISTS idx_monster_templates_genre_id;

ALTER TABLE monsters
  DROP COLUMN IF EXISTS genre_id;

ALTER TABLE monster_templates
  DROP COLUMN IF EXISTS genre_id;

ALTER TABLE zone_genres
  DROP COLUMN IF EXISTS prompt_seed;
