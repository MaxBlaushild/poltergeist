CREATE TABLE resource_gather_requirements_old (
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

INSERT INTO resource_gather_requirements_old (
  created_at,
  updated_at,
  resource_id,
  min_level,
  max_level,
  required_inventory_item_id
)
SELECT
  COALESCE(rgr.created_at, NOW()) AS created_at,
  COALESCE(rgr.updated_at, NOW()) AS updated_at,
  r.id AS resource_id,
  rgr.min_level,
  rgr.max_level,
  rgr.required_inventory_item_id
FROM resource_gather_requirements rgr
JOIN resources r
  ON r.resource_type_id = rgr.resource_type_id;

DROP TABLE resource_gather_requirements;

ALTER TABLE resource_gather_requirements_old
  RENAME TO resource_gather_requirements;

CREATE INDEX idx_resource_gather_requirements_resource_id
  ON resource_gather_requirements(resource_id);

CREATE INDEX idx_resource_gather_requirements_required_inventory_item_id
  ON resource_gather_requirements(required_inventory_item_id);
