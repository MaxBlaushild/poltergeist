DROP TABLE IF EXISTS quest_archetype_spell_rewards;

ALTER TABLE quest_archetypes
DROP COLUMN IF EXISTS character_tags,
DROP COLUMN IF EXISTS material_rewards_json,
DROP COLUMN IF EXISTS recurrence_frequency,
DROP COLUMN IF EXISTS reward_experience,
DROP COLUMN IF EXISTS random_reward_size,
DROP COLUMN IF EXISTS reward_mode,
DROP COLUMN IF EXISTS image_url,
DROP COLUMN IF EXISTS acceptance_dialogue,
DROP COLUMN IF EXISTS description;
