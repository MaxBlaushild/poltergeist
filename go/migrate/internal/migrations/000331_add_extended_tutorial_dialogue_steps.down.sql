ALTER TABLE tutorial_configs
  DROP COLUMN IF EXISTS post_base_dialogue_json,
  DROP COLUMN IF EXISTS base_kit_dialogue_json,
  DROP COLUMN IF EXISTS post_monster_dialogue_json;
