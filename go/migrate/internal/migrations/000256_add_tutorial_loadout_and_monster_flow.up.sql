ALTER TABLE tutorial_configs
  ADD COLUMN IF NOT EXISTS loadout_dialogue_json JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS monster_encounter_id UUID REFERENCES monster_encounters(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS monster_reward_experience INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS monster_reward_gold INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS monster_item_rewards_json JSONB NOT NULL DEFAULT '[]';

UPDATE tutorial_configs
SET loadout_dialogue_json = '["Equip your new gear and use the spellbook before you head back out."]'
WHERE loadout_dialogue_json = '[]'::jsonb;

ALTER TABLE user_tutorial_states
  ADD COLUMN IF NOT EXISTS stage TEXT NOT NULL DEFAULT 'welcome',
  ADD COLUMN IF NOT EXISTS selected_scenario_option_id UUID REFERENCES scenario_options(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS required_equip_item_ids_json JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS completed_equip_item_ids_json JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS required_use_item_ids_json JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS completed_use_item_ids_json JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS tutorial_monster_encounter_id UUID REFERENCES monster_encounters(id) ON DELETE SET NULL;

UPDATE user_tutorial_states
SET stage = CASE
  WHEN completed_at IS NOT NULL THEN 'completed'
  WHEN tutorial_scenario_id IS NOT NULL THEN 'scenario'
  ELSE 'welcome'
END
WHERE stage IS NULL OR BTRIM(stage) = '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_tutorial_states_monster_encounter_id
  ON user_tutorial_states(tutorial_monster_encounter_id)
  WHERE tutorial_monster_encounter_id IS NOT NULL;

ALTER TABLE monsters
  ADD COLUMN IF NOT EXISTS owner_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS ephemeral BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_monsters_owner_user_id ON monsters(owner_user_id);

ALTER TABLE monster_encounters
  ADD COLUMN IF NOT EXISTS owner_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS ephemeral BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_monster_encounters_owner_user_id ON monster_encounters(owner_user_id);
