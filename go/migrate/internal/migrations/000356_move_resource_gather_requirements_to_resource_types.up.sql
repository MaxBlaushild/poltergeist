CREATE TABLE resource_gather_requirements_new (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  resource_type_id UUID NOT NULL REFERENCES resource_types(id) ON DELETE CASCADE,
  min_level INTEGER NOT NULL,
  max_level INTEGER NOT NULL,
  required_inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
  CONSTRAINT resource_gather_requirements_min_level_positive CHECK (min_level > 0),
  CONSTRAINT resource_gather_requirements_max_level_positive CHECK (max_level > 0),
  CONSTRAINT resource_gather_requirements_level_range_valid CHECK (max_level >= min_level)
);

INSERT INTO resource_gather_requirements_new (
  created_at,
  updated_at,
  resource_type_id,
  min_level,
  max_level,
  required_inventory_item_id
)
SELECT
  COALESCE(MIN(rgr.created_at), NOW()) AS created_at,
  COALESCE(MAX(rgr.updated_at), NOW()) AS updated_at,
  r.resource_type_id,
  rgr.min_level,
  rgr.max_level,
  rgr.required_inventory_item_id
FROM resource_gather_requirements rgr
JOIN resources r
  ON r.id = rgr.resource_id
GROUP BY
  r.resource_type_id,
  rgr.min_level,
  rgr.max_level,
  rgr.required_inventory_item_id;

DROP TABLE resource_gather_requirements;

ALTER TABLE resource_gather_requirements_new
  RENAME TO resource_gather_requirements;

CREATE INDEX idx_resource_gather_requirements_resource_type_id
  ON resource_gather_requirements(resource_type_id);

CREATE INDEX idx_resource_gather_requirements_required_inventory_item_id
  ON resource_gather_requirements(required_inventory_item_id);
