-- Same bug 000436 fixed for magnetic-frag-rack: widthMm/depthMm had no
-- "default" key, so the frontend (which falls back to each field's bare
-- `minimum` when no default is set) landed every first-time visitor on
-- widthMm=60/depthMm=40 — a combination that can't fit holesPerRow's own
-- schema minimum of 4 (only 2 holes fit at 60mm width with 20mm plug
-- holes). The client-side derived-bound clamp then fought itself: it
-- raised widthMm to fit 4 holes in the same pass it had already clamped
-- holesPerRow down to 2 using the *old* (too-narrow) width, landing on
-- widthMm=95/holesPerRow=2 — which ValidateParams correctly rejects since
-- holesPerRow's static schema minimum is 4. Every field was individually
-- within its own min/max; nothing was wrong with the ranges, just the
-- implied default combination.
--
-- widthMm=130/depthMm=60/legHeightMm=30/holesPerRow=4/rowCount=1/
-- plugHoleDiameterMm=20 verified against both ShelfRack.ValidateParams and
-- Analyze (MinWallMm=4mm, the deck's own thickness — comfortably above the
-- 2mm floor) before writing this, with slack: ShelfRackMaxHolesPerRow(130,
-- 20)=5 and ShelfRackMaxRows(60,20)=2, so the defaults aren't sitting
-- exactly on the derived boundary either.

UPDATE reef_parameter_schemas
SET active = false
WHERE product_id = (SELECT id FROM reef_products WHERE slug = 'shelf-rack')
  AND generator_module = 'shelf_rack'
  AND version = 1;

INSERT INTO reef_parameter_schemas (product_id, version, schema, generator_module, generator_version, active)
SELECT id, 2, $shelf$
{
  "type": "object",
  "required": ["widthMm", "depthMm", "legHeightMm", "plugHoleDiameterMm", "holesPerRow", "rowCount", "color"],
  "properties": {
    "widthMm": {
      "type": "number",
      "minimum": 60,
      "maximum": 250,
      "default": 130,
      "x-unit": "mm",
      "x-label": "Deck width",
      "x-helpText": "Measure the footprint width where the rack will stand. Width also caps how many holes fit per row.",
      "x-diagramAsset": "/reef/diagrams/shelf-width.svg"
    },
    "depthMm": {
      "type": "number",
      "minimum": 40,
      "maximum": 150,
      "default": 60,
      "x-unit": "mm",
      "x-label": "Deck depth",
      "x-helpText": "Measure the footprint depth where the rack will stand. Depth also caps how many rows of holes fit.",
      "x-diagramAsset": "/reef/diagrams/shelf-depth.svg"
    },
    "legHeightMm": {
      "type": "number",
      "minimum": 15,
      "maximum": 100,
      "default": 30,
      "x-unit": "mm",
      "x-label": "Leg height",
      "x-helpText": "How far the deck stands above the tank floor, sump shelf, or rockwork it rests on.",
      "x-diagramAsset": "/reef/diagrams/shelf-leg-height.svg"
    },
    "plugHoleDiameterMm": {
      "type": "integer",
      "enum": [15, 20],
      "default": 20,
      "x-unit": "mm",
      "x-label": "Frag plug hole diameter",
      "x-helpText": "Standard frag plug stems are 15mm or 20mm. Measure your plug stem diameter, not the plug head.",
      "x-diagramAsset": "/reef/diagrams/plug-hole-diameter.svg"
    },
    "holesPerRow": {
      "type": "integer",
      "minimum": 4,
      "maximum": 12,
      "default": 4,
      "x-label": "Holes per row",
      "x-helpText": "How many frag plugs per row. The maximum is derived from deck width and plug hole diameter so holes never overlap.",
      "x-derivedBoundFrom": ["widthMm", "plugHoleDiameterMm"]
    },
    "rowCount": {
      "type": "integer",
      "minimum": 1,
      "maximum": 4,
      "default": 1,
      "x-label": "Rows",
      "x-helpText": "Rows of frag-plug holes across the deck's depth. The maximum is derived from deck depth and plug hole diameter so holes never overlap.",
      "x-derivedBoundFrom": ["depthMm", "plugHoleDiameterMm"]
    },
    "color": {
      "type": "string",
      "enum": ["black", "white", "gray", "clear", "blue"],
      "default": "black",
      "x-label": "Color",
      "x-helpText": "PETG filament color. Black is the default."
    }
  }
}
$shelf$::jsonb, 'shelf_rack', 'v1', true
FROM reef_products WHERE slug = 'shelf-rack'
ON CONFLICT (product_id, version) DO NOTHING;
