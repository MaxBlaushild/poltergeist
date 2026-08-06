-- Adds a third mounting style alongside magnetic-frag-rack (glass clips)
-- and lid-mesh-clips (rim clips): a freestanding deck standing on four
-- printed legs/stilts, for tanks/sumps/shelves with no glass to clip a
-- magnetic rack to. Prices entirely from reef_slice_results.price_cents
-- like every other configurable product (R-6.2), so base_price_cents is 0.

INSERT INTO reef_products (slug, name, kind, description, material, base_price_cents, images, active) VALUES
  ('shelf-rack', 'Shelf Frag Rack', 'configurable',
   'A made-to-order freestanding frag rack that stands on printed legs instead of clipping to glass — sits on the tank floor, sump shelf, or rockwork.',
   'PETG', 0, '[]'::jsonb, true)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO reef_parameter_schemas (product_id, version, schema, generator_module, generator_version, active)
SELECT id, 1, $shelf$
{
  "type": "object",
  "required": ["widthMm", "depthMm", "legHeightMm", "plugHoleDiameterMm", "holesPerRow", "rowCount", "color"],
  "properties": {
    "widthMm": {
      "type": "number",
      "minimum": 60,
      "maximum": 250,
      "x-unit": "mm",
      "x-label": "Deck width",
      "x-helpText": "Measure the footprint width where the rack will stand. Width also caps how many holes fit per row.",
      "x-diagramAsset": "/reef/diagrams/shelf-width.svg"
    },
    "depthMm": {
      "type": "number",
      "minimum": 40,
      "maximum": 150,
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
