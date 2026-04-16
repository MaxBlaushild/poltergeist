CREATE TABLE IF NOT EXISTS resource_gather_requirements (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  resource_id UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  min_level INTEGER NOT NULL,
  max_level INTEGER NOT NULL,
  required_inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
  CONSTRAINT resource_gather_requirements_min_level_positive CHECK (min_level > 0),
  CONSTRAINT resource_gather_requirements_max_level_positive CHECK (max_level > 0),
  CONSTRAINT resource_gather_requirements_level_range_valid CHECK (max_level >= min_level)
);

CREATE INDEX IF NOT EXISTS idx_resource_gather_requirements_resource_id
  ON resource_gather_requirements(resource_id);

CREATE INDEX IF NOT EXISTS idx_resource_gather_requirements_required_inventory_item_id
  ON resource_gather_requirements(required_inventory_item_id);
