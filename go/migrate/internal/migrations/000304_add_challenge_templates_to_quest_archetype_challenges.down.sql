DROP INDEX IF EXISTS idx_quest_archetype_challenges_challenge_template_id;

ALTER TABLE quest_archetype_challenges
  DROP COLUMN IF EXISTS challenge_template_id;
