ALTER TABLE quest_acceptances_v2
  DROP COLUMN IF EXISTS objectives_completed_at,
  DROP COLUMN IF EXISTS closed_at,
  DROP COLUMN IF EXISTS closure_method,
  DROP COLUMN IF EXISTS debrief_pending,
  DROP COLUMN IF EXISTS debriefed_at;

ALTER TABLE quests
  DROP COLUMN IF EXISTS closure_policy,
  DROP COLUMN IF EXISTS debrief_policy,
  DROP COLUMN IF EXISTS return_bonus_gold,
  DROP COLUMN IF EXISTS return_bonus_experience,
  DROP COLUMN IF EXISTS return_bonus_relationship_effects;

ALTER TABLE quest_archetypes
  DROP COLUMN IF EXISTS closure_policy,
  DROP COLUMN IF EXISTS debrief_policy,
  DROP COLUMN IF EXISTS return_bonus_gold,
  DROP COLUMN IF EXISTS return_bonus_experience,
  DROP COLUMN IF EXISTS return_bonus_relationship_effects;
