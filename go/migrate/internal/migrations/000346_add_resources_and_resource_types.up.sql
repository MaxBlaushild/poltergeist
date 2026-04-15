CREATE TABLE IF NOT EXISTS resource_types (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  map_icon_url TEXT NOT NULL DEFAULT '',
  map_icon_prompt TEXT NOT NULL DEFAULT ''
);

INSERT INTO resource_types (name, slug, description)
SELECT 'Herbalism', 'herbalism', 'Gatherable herbs, flowers, and wild plants.'
WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE slug = 'herbalism');

INSERT INTO resource_types (name, slug, description)
SELECT 'Mining', 'mining', 'Ore veins, stone nodes, and mineral deposits.'
WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE slug = 'mining');

INSERT INTO resource_types (name, slug, description)
SELECT 'Logging', 'logging', 'Trees, timber, and chopped wood sources.'
WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE slug = 'logging');

INSERT INTO resource_types (name, slug, description)
SELECT 'Skinning', 'skinning', 'Creature remains and harvestable hides.'
WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE slug = 'skinning');

INSERT INTO resource_types (name, slug, description)
SELECT 'Fishing', 'fishing', 'Fishing holes, shoals, and waterside catches.'
WHERE NOT EXISTS (SELECT 1 FROM resource_types WHERE slug = 'fishing');

ALTER TABLE inventory_items
  ADD COLUMN IF NOT EXISTS resource_type_id UUID;

UPDATE inventory_items AS ii
SET resource_type_id = rt.id
FROM resource_types AS rt
WHERE ii.resource_type_id IS NULL
  AND LOWER(TRIM(COALESCE(ii.resource_type, ''))) = rt.slug;

ALTER TABLE inventory_items
  DROP CONSTRAINT IF EXISTS inventory_items_resource_type_id_fkey;

ALTER TABLE inventory_items
  ADD CONSTRAINT inventory_items_resource_type_id_fkey
  FOREIGN KEY (resource_type_id)
  REFERENCES resource_types(id)
  ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_inventory_items_resource_type_id
  ON inventory_items(resource_type_id);

ALTER TABLE inventory_items
  DROP CONSTRAINT IF EXISTS inventory_items_resource_type_check;

ALTER TABLE inventory_items
  DROP COLUMN IF EXISTS resource_type;

CREATE TABLE IF NOT EXISTS resources (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  resource_type_id UUID NOT NULL REFERENCES resource_types(id) ON DELETE RESTRICT,
  inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id) ON DELETE RESTRICT,
  quantity INTEGER NOT NULL DEFAULT 1,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  geometry geometry(Point,4326),
  invalidated BOOLEAN NOT NULL DEFAULT FALSE,
  CONSTRAINT resources_quantity_positive CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_resources_zone_id
  ON resources(zone_id);

CREATE INDEX IF NOT EXISTS idx_resources_resource_type_id
  ON resources(resource_type_id);

CREATE INDEX IF NOT EXISTS idx_resources_inventory_item_id
  ON resources(inventory_item_id);

CREATE INDEX IF NOT EXISTS idx_resources_geometry
  ON resources USING GIST (geometry);

CREATE TABLE IF NOT EXISTS user_resource_gatherings (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  resource_id UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  gathered_at TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT user_resource_gatherings_user_resource_unique UNIQUE (user_id, resource_id)
);

CREATE INDEX IF NOT EXISTS idx_user_resource_gatherings_user_id
  ON user_resource_gatherings(user_id);

CREATE INDEX IF NOT EXISTS idx_user_resource_gatherings_resource_id
  ON user_resource_gatherings(resource_id);
