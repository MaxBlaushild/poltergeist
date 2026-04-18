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

ALTER TABLE scenarios
  ADD COLUMN genre_id UUID;

ALTER TABLE scenario_templates
  ADD COLUMN genre_id UUID;

ALTER TABLE scenario_generation_jobs
  ADD COLUMN genre_id UUID;

ALTER TABLE scenario_template_generation_jobs
  ADD COLUMN genre_id UUID;

UPDATE scenarios
SET genre_id = fantasy.id
FROM (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC
  LIMIT 1
) fantasy
WHERE scenarios.genre_id IS NULL;

UPDATE scenario_templates
SET genre_id = fantasy.id
FROM (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC
  LIMIT 1
) fantasy
WHERE scenario_templates.genre_id IS NULL;

UPDATE scenario_generation_jobs
SET genre_id = fantasy.id
FROM (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC
  LIMIT 1
) fantasy
WHERE scenario_generation_jobs.genre_id IS NULL;

UPDATE scenario_template_generation_jobs
SET genre_id = fantasy.id
FROM (
  SELECT id
  FROM zone_genres
  WHERE LOWER(name) = LOWER('Fantasy')
  ORDER BY sort_order ASC, created_at ASC
  LIMIT 1
) fantasy
WHERE scenario_template_generation_jobs.genre_id IS NULL;

ALTER TABLE scenarios
  ALTER COLUMN genre_id SET NOT NULL,
  ADD CONSTRAINT scenarios_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES zone_genres(id);

ALTER TABLE scenario_templates
  ALTER COLUMN genre_id SET NOT NULL,
  ADD CONSTRAINT scenario_templates_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES zone_genres(id);

ALTER TABLE scenario_generation_jobs
  ALTER COLUMN genre_id SET NOT NULL,
  ADD CONSTRAINT scenario_generation_jobs_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES zone_genres(id);

ALTER TABLE scenario_template_generation_jobs
  ALTER COLUMN genre_id SET NOT NULL,
  ADD CONSTRAINT scenario_template_generation_jobs_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES zone_genres(id);

CREATE INDEX idx_scenarios_genre_id ON scenarios(genre_id);
CREATE INDEX idx_scenario_templates_genre_id ON scenario_templates(genre_id);
CREATE INDEX idx_scenario_generation_jobs_genre_id ON scenario_generation_jobs(genre_id);
CREATE INDEX idx_scenario_template_generation_jobs_genre_id ON scenario_template_generation_jobs(genre_id);

COMMIT;
