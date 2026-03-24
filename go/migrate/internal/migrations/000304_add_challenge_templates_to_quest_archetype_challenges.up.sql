ALTER TABLE quest_archetype_challenges
  ADD COLUMN IF NOT EXISTS challenge_template_id UUID REFERENCES challenge_templates(id);

CREATE INDEX IF NOT EXISTS idx_quest_archetype_challenges_challenge_template_id
  ON quest_archetype_challenges(challenge_template_id);
