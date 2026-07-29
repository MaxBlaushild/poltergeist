-- widthMm, tierCount, holesPerTier, and glassThicknessMm had no "default"
-- key, so the frontend (which falls back to each field's bare `minimum`
-- when no default is set) landed every first-time visitor on
-- widthMm=60/tierCount=1/holesPerTier=4/glassThicknessMm=4 — a combination
-- that FragRack.ValidateParams (go/pkg/reef/generate/frag_rack.go) now
-- correctly rejects on the very first preview: only 2 holes fit at 60mm
-- width with 20mm plug holes, not 4. Every field was individually within
-- its own min/max; nothing was wrong with the ranges, just the implied
-- default combination.
--
-- widthMm=150/tierCount=2/holesPerTier=5/glassThicknessMm=8 verified
-- against both FragRack.ValidateParams and Analyze (MinWallMm=2.7mm,
-- comfortably above the 2.0mm minimum) before writing this.

UPDATE reef_parameter_schemas
SET active = false
WHERE product_id = (SELECT id FROM reef_products WHERE slug = 'magnetic-frag-rack')
  AND generator_module = 'frag_rack'
  AND version = 2;

INSERT INTO reef_parameter_schemas (product_id, version, schema, generator_module, generator_version, active)
SELECT id, 3, $frag$
{
  "type": "object",
  "required": ["glassThicknessMm", "tierCount", "widthMm", "plugHoleDiameterMm", "holesPerTier", "color"],
  "properties": {
    "tankProfileId": {
      "type": ["string", "null"],
      "x-control": "tank-select",
      "x-label": "Tank model",
      "x-helpText": "Pick your tank model to auto-fill glass thickness. Not listed? Choose \"Other\" and measure by hand.",
      "x-diagramAsset": "/reef/diagrams/tank-select.svg",
      "x-autofills": ["glassThicknessMm"]
    },
    "glassThicknessMm": {
      "type": "number",
      "minimum": 4,
      "maximum": 19,
      "default": 8,
      "x-unit": "mm",
      "x-label": "Glass thickness",
      "x-helpText": "Measure straight across the glass edge at the rim with calipers. Above 19mm the magnets in this design cannot hold reliably, so the range stops there.",
      "x-diagramAsset": "/reef/diagrams/glass-thickness.svg"
    },
    "tierCount": {
      "type": "integer",
      "minimum": 1,
      "maximum": 4,
      "default": 2,
      "x-label": "Tiers",
      "x-helpText": "Number of stacked frag-plug tiers. More tiers means more magnet pairs to hold the rack's weight.",
      "x-diagramAsset": "/reef/diagrams/tier-count.svg"
    },
    "widthMm": {
      "type": "number",
      "minimum": 60,
      "maximum": 250,
      "default": 150,
      "x-unit": "mm",
      "x-label": "Rack width",
      "x-helpText": "Measure the usable rim length where the rack will hang. Width also caps how many holes fit per tier.",
      "x-diagramAsset": "/reef/diagrams/rack-width.svg"
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
    "holesPerTier": {
      "type": "integer",
      "minimum": 4,
      "maximum": 12,
      "default": 5,
      "x-label": "Holes per tier",
      "x-helpText": "How many frag plugs per tier. The maximum is derived from rack width and plug hole diameter so holes never overlap.",
      "x-derivedBoundFrom": ["widthMm", "plugHoleDiameterMm"]
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
$frag$::jsonb, 'frag_rack', 'v1', true
FROM reef_products WHERE slug = 'magnetic-frag-rack'
ON CONFLICT (product_id, version) DO NOTHING;
