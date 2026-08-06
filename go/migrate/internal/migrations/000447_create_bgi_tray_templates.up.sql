-- One row per hand-designed tray template (R-1.1: curated-parametric, not
-- generative, for v1). generator_module matches a
-- go/pkg/reef/generate.Module.Slug() the set-assembly service resolves at
-- runtime.
CREATE TABLE IF NOT EXISTS bgi_tray_templates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  slug TEXT NOT NULL UNIQUE,
  generator_module TEXT NOT NULL,
  generator_version TEXT NOT NULL,
  component_type TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_bgi_tray_templates_component_type ON bgi_tray_templates(component_type);
