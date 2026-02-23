CREATE TABLE quest_generation_jobs (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  zone_quest_archetype_id uuid NOT NULL REFERENCES zone_quest_archetypes(id) ON DELETE CASCADE,
  zone_id uuid NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  quest_archetype_id uuid NOT NULL REFERENCES quest_archetypes(id) ON DELETE CASCADE,
  quest_giver_character_id uuid NULL REFERENCES characters(id) ON DELETE SET NULL,
  status text NOT NULL DEFAULT 'queued',
  total_count int NOT NULL DEFAULT 0,
  completed_count int NOT NULL DEFAULT 0,
  failed_count int NOT NULL DEFAULT 0,
  error_message text,
  quest_ids jsonb NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX quest_generation_jobs_zone_quest_archetype_id_idx
  ON quest_generation_jobs(zone_quest_archetype_id);

CREATE INDEX quest_generation_jobs_created_at_idx
  ON quest_generation_jobs(created_at DESC);
