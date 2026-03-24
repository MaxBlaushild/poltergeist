ALTER TABLE quest_archetype_nodes
  ADD COLUMN IF NOT EXISTS scenario_template_id UUID REFERENCES scenario_templates(id) ON DELETE SET NULL;
