CREATE TABLE IF NOT EXISTS inventory_item_suggestion_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  status TEXT NOT NULL,
  count INTEGER NOT NULL DEFAULT 1,
  theme_prompt TEXT NOT NULL DEFAULT '',
  categories JSONB NOT NULL DEFAULT '[]'::jsonb,
  rarity_tiers JSONB NOT NULL DEFAULT '[]'::jsonb,
  equip_slots JSONB NOT NULL DEFAULT '[]'::jsonb,
  internal_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  min_item_level INTEGER NOT NULL DEFAULT 1,
  max_item_level INTEGER NOT NULL DEFAULT 100,
  created_count INTEGER NOT NULL DEFAULT 0,
  error_message TEXT
);

CREATE INDEX IF NOT EXISTS inventory_item_suggestion_jobs_created_at_idx
  ON inventory_item_suggestion_jobs(created_at DESC);

CREATE TABLE IF NOT EXISTS inventory_item_suggestion_drafts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  job_id UUID NOT NULL REFERENCES inventory_item_suggestion_jobs(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'suggested',
  name TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT '',
  rarity_tier TEXT NOT NULL DEFAULT 'Common',
  item_level INTEGER NOT NULL DEFAULT 1,
  equip_slot TEXT,
  why_it_fits TEXT NOT NULL DEFAULT '',
  internal_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  warnings JSONB NOT NULL DEFAULT '[]'::jsonb,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  inventory_item_id INTEGER REFERENCES inventory_items(id) ON DELETE SET NULL,
  converted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS inventory_item_suggestion_drafts_job_id_idx
  ON inventory_item_suggestion_drafts(job_id, created_at DESC);

CREATE INDEX IF NOT EXISTS inventory_item_suggestion_drafts_status_idx
  ON inventory_item_suggestion_drafts(status);
