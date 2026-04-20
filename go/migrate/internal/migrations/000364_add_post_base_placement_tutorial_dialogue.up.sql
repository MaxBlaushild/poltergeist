ALTER TABLE tutorial_configs
  ADD COLUMN IF NOT EXISTS post_base_placement_dialogue_json JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE tutorial_configs
SET post_base_placement_dialogue_json = COALESCE(post_base_placement_dialogue_json, '[]'::jsonb)
WHERE post_base_placement_dialogue_json IS NULL;
