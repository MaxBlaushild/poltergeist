ALTER TABLE quest_archetype_nodes
  ADD COLUMN IF NOT EXISTS encounter_type TEXT NOT NULL DEFAULT 'monster';
