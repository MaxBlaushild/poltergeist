DROP INDEX IF EXISTS idx_monster_encounters_owner_user_id;
ALTER TABLE monster_encounters
  DROP COLUMN IF EXISTS ephemeral,
  DROP COLUMN IF EXISTS owner_user_id;

DROP INDEX IF EXISTS idx_monsters_owner_user_id;
ALTER TABLE monsters
  DROP COLUMN IF EXISTS ephemeral,
  DROP COLUMN IF EXISTS owner_user_id;

DROP INDEX IF EXISTS idx_user_tutorial_states_monster_encounter_id;
ALTER TABLE user_tutorial_states
  DROP COLUMN IF EXISTS tutorial_monster_encounter_id,
  DROP COLUMN IF EXISTS completed_use_item_ids_json,
  DROP COLUMN IF EXISTS required_use_item_ids_json,
  DROP COLUMN IF EXISTS completed_equip_item_ids_json,
  DROP COLUMN IF EXISTS required_equip_item_ids_json,
  DROP COLUMN IF EXISTS selected_scenario_option_id,
  DROP COLUMN IF EXISTS stage;

ALTER TABLE tutorial_configs
  DROP COLUMN IF EXISTS monster_item_rewards_json,
  DROP COLUMN IF EXISTS monster_reward_gold,
  DROP COLUMN IF EXISTS monster_reward_experience,
  DROP COLUMN IF EXISTS monster_encounter_id,
  DROP COLUMN IF EXISTS loadout_dialogue_json;
