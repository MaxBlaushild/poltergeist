ALTER TABLE quest_archetypes
  ADD COLUMN IF NOT EXISTS closure_policy TEXT NOT NULL DEFAULT 'remote',
  ADD COLUMN IF NOT EXISTS debrief_policy TEXT NOT NULL DEFAULT 'optional',
  ADD COLUMN IF NOT EXISTS return_bonus_gold INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS return_bonus_experience INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS return_bonus_relationship_effects JSONB NOT NULL DEFAULT '{}';

ALTER TABLE quests
  ADD COLUMN IF NOT EXISTS closure_policy TEXT NOT NULL DEFAULT 'remote',
  ADD COLUMN IF NOT EXISTS debrief_policy TEXT NOT NULL DEFAULT 'optional',
  ADD COLUMN IF NOT EXISTS return_bonus_gold INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS return_bonus_experience INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS return_bonus_relationship_effects JSONB NOT NULL DEFAULT '{}';

ALTER TABLE quest_acceptances_v2
  ADD COLUMN IF NOT EXISTS objectives_completed_at TIMESTAMP,
  ADD COLUMN IF NOT EXISTS closed_at TIMESTAMP,
  ADD COLUMN IF NOT EXISTS closure_method TEXT NOT NULL DEFAULT 'in_person',
  ADD COLUMN IF NOT EXISTS debrief_pending BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS debriefed_at TIMESTAMP;

UPDATE quest_archetypes
SET
  closure_policy = CASE
    WHEN category = 'main_story' THEN 'in_person'
    ELSE 'remote'
  END,
  debrief_policy = CASE
    WHEN category = 'main_story' THEN 'required_for_followup'
    ELSE 'optional'
  END
WHERE TRUE;

UPDATE quests
SET
  closure_policy = CASE
    WHEN category = 'main_story' THEN 'in_person'
    ELSE 'remote'
  END,
  debrief_policy = CASE
    WHEN category = 'main_story' THEN 'required_for_followup'
    ELSE 'optional'
  END
WHERE TRUE;

UPDATE quest_acceptances_v2
SET
  objectives_completed_at = COALESCE(objectives_completed_at, turned_in_at),
  closed_at = COALESCE(closed_at, turned_in_at),
  closure_method = CASE
    WHEN turned_in_at IS NOT NULL THEN 'in_person'
    ELSE closure_method
  END,
  debrief_pending = CASE
    WHEN turned_in_at IS NOT NULL THEN FALSE
    ELSE debrief_pending
  END,
  debriefed_at = COALESCE(debriefed_at, turned_in_at)
WHERE turned_in_at IS NOT NULL;
