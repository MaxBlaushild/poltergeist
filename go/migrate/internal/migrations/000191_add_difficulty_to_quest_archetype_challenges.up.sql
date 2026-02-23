ALTER TABLE quest_archetype_challenges
  ADD COLUMN IF NOT EXISTS difficulty INTEGER NOT NULL DEFAULT 0;
