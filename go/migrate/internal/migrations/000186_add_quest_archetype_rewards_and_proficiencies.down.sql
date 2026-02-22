DROP TABLE IF EXISTS quest_archetype_item_rewards;

ALTER TABLE quest_archetype_challenges
  DROP COLUMN IF EXISTS inventory_item_id,
  DROP COLUMN IF EXISTS proficiency;
