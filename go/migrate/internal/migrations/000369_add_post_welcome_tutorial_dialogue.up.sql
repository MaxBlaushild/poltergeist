ALTER TABLE tutorial_configs
  ADD COLUMN IF NOT EXISTS post_welcome_dialogue_json JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE tutorial_configs
SET post_welcome_dialogue_json = COALESCE(post_welcome_dialogue_json, '[]'::jsonb)
WHERE post_welcome_dialogue_json IS NULL;
