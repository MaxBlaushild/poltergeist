CREATE TABLE IF NOT EXISTS user_story_flags (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  flag_key text NOT NULL,
  value boolean NOT NULL DEFAULT true
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_story_flags_user_key
ON user_story_flags (user_id, flag_key);

ALTER TABLE quests
ADD COLUMN IF NOT EXISTS required_story_flags jsonb NOT NULL DEFAULT '[]'::jsonb,
ADD COLUMN IF NOT EXISTS set_story_flags jsonb NOT NULL DEFAULT '[]'::jsonb,
ADD COLUMN IF NOT EXISTS clear_story_flags jsonb NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE quest_archetypes
ADD COLUMN IF NOT EXISTS required_story_flags jsonb NOT NULL DEFAULT '[]'::jsonb,
ADD COLUMN IF NOT EXISTS set_story_flags jsonb NOT NULL DEFAULT '[]'::jsonb,
ADD COLUMN IF NOT EXISTS clear_story_flags jsonb NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE characters
ADD COLUMN IF NOT EXISTS story_variants jsonb NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE points_of_interest
ADD COLUMN IF NOT EXISTS story_variants jsonb NOT NULL DEFAULT '[]'::jsonb;
