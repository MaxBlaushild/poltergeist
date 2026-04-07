ALTER TABLE tutorial_configs
  ADD COLUMN IF NOT EXISTS post_monster_dialogue_json JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS base_kit_dialogue_json JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS post_base_dialogue_json JSONB NOT NULL DEFAULT '[]';

UPDATE tutorial_configs
SET base_kit_dialogue_json = '["Use the home base kit you just earned and claim a safe place for yourself."]'
WHERE base_kit_dialogue_json = '[]'::jsonb;
