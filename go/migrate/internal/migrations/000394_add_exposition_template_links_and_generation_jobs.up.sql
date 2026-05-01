ALTER TABLE exposition_templates
  ADD COLUMN IF NOT EXISTS zone_kind TEXT NOT NULL DEFAULT '';

ALTER TABLE expositions
  ADD COLUMN IF NOT EXISTS exposition_template_id UUID REFERENCES exposition_templates(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_expositions_exposition_template_id
  ON expositions(exposition_template_id);

ALTER TABLE zone_seed_jobs
  ADD COLUMN IF NOT EXISTS exposition_count INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS exposition_template_generation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  zone_kind TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'queued',
  count INTEGER NOT NULL DEFAULT 1,
  created_count INTEGER NOT NULL DEFAULT 0,
  error_message TEXT
);

CREATE INDEX IF NOT EXISTS exposition_template_generation_jobs_created_at_idx
  ON exposition_template_generation_jobs(created_at DESC);

CREATE INDEX IF NOT EXISTS exposition_template_generation_jobs_zone_kind_idx
  ON exposition_template_generation_jobs(zone_kind);

CREATE TEMP TABLE exposition_template_backfill_map (
  exposition_id UUID PRIMARY KEY,
  template_id UUID NOT NULL
) ON COMMIT DROP;

INSERT INTO exposition_template_backfill_map (exposition_id, template_id)
SELECT e.id, uuid_generate_v4()
FROM expositions e
WHERE e.exposition_template_id IS NULL;

WITH source AS (
  SELECT
    e.id AS exposition_id,
    m.template_id,
    e.created_at,
    e.updated_at,
    COALESCE(e.zone_kind, '') AS zone_kind,
    e.title,
    e.description,
    e.dialogue,
    e.required_story_flags,
    e.image_url,
    e.thumbnail_url,
    e.reward_mode,
    e.random_reward_size,
    e.reward_experience,
    e.reward_gold,
    e.material_rewards_json,
    COALESCE(
      (
        SELECT jsonb_agg(
          jsonb_build_object(
            'inventoryItemId',
            eir.inventory_item_id,
            'quantity',
            eir.quantity
          )
          ORDER BY eir.created_at, eir.id
        )
        FROM exposition_item_rewards eir
        WHERE eir.exposition_id = e.id
      ),
      '[]'::jsonb
    ) AS item_rewards_json,
    COALESCE(
      (
        SELECT jsonb_agg(
          jsonb_build_object('spellId', esr.spell_id)
          ORDER BY esr.created_at, esr.id
        )
        FROM exposition_spell_rewards esr
        WHERE esr.exposition_id = e.id
      ),
      '[]'::jsonb
    ) AS spell_rewards_json
  FROM expositions e
  JOIN exposition_template_backfill_map m
    ON m.exposition_id = e.id
)
INSERT INTO exposition_templates (
  id,
  created_at,
  updated_at,
  zone_kind,
  title,
  description,
  dialogue,
  required_story_flags,
  image_url,
  thumbnail_url,
  reward_mode,
  random_reward_size,
  reward_experience,
  reward_gold,
  material_rewards_json,
  item_rewards_json,
  spell_rewards_json
)
SELECT
  template_id,
  created_at,
  updated_at,
  zone_kind,
  title,
  description,
  dialogue,
  required_story_flags,
  image_url,
  thumbnail_url,
  reward_mode,
  random_reward_size,
  reward_experience,
  reward_gold,
  material_rewards_json,
  item_rewards_json,
  spell_rewards_json
FROM source;

UPDATE expositions e
SET exposition_template_id = m.template_id
FROM exposition_template_backfill_map m
WHERE e.id = m.exposition_id
  AND e.exposition_template_id IS NULL;
