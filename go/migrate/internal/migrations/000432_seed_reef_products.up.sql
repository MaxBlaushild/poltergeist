-- R-1.1 catalog: two configurables + four fixed SKUs.
-- base_price_cents / variant price_cents below are [DECIDE] placeholders
-- (R-6.1: "derive them from real fulfillment quotes, not planning estimates").
-- Configurable products price entirely from reef_slice_results.price_cents
-- (R-6.2), so their base_price_cents is 0 and unused by the pricing formula;
-- it exists only as a schema-required column.

INSERT INTO reef_products (slug, name, kind, description, material, base_price_cents, images, active) VALUES
  ('magnetic-frag-rack', 'Magnetic Frag Rack', 'configurable',
   'A made-to-order magnetic frag rack, parametrically fit to your tank''s rim and glass thickness.',
   'PETG', 0, '[]'::jsonb, true),
  ('lid-mesh-clips', 'Lid & Mesh Screen Clips (Set)', 'configurable',
   'Made-to-order clips that hold your lid or mesh screen to the tank rim, sized to your glass and rim.',
   'PETG', 0, '[]'::jsonb, true),
  ('feeding-ring', 'Feeding Ring', 'fixed',
   'A floating feeding ring that keeps frozen and pellet food from scattering across the surface.',
   'PETG', 1400, '[]'::jsonb, true),
  ('dosing-tube-organizer', 'Dosing Tube Organizer', 'fixed',
   'Keeps dosing pump tubing organized and untangled behind the tank.',
   'PETG', 1800, '[]'::jsonb, true),
  ('ato-float-switch-bracket', 'ATO Float-Switch Bracket', 'fixed',
   'A bracket that mounts a float switch at a fixed, repeatable height in the sump.',
   'PETG', 1000, '[]'::jsonb, true),
  ('frag-plug-tray', 'Frag Plug Tray — 24-Hole', 'fixed',
   'A 24-hole tray for organizing frag plugs in the frag zone.',
   'PETG', 2200, '[]'::jsonb, true)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO reef_product_variants (product_id, variant_key, label, price_cents, active)
SELECT id, v.variant_key, v.label, v.price_cents, true
FROM reef_products, LATERAL (VALUES
  ('small', 'Small (60mm)', 1400),
  ('medium', 'Medium (90mm)', 1600),
  ('large', 'Large (120mm)', 1800)
) AS v(variant_key, label, price_cents)
WHERE reef_products.slug = 'feeding-ring'
ON CONFLICT (product_id, variant_key) DO NOTHING;

INSERT INTO reef_product_variants (product_id, variant_key, label, price_cents, active)
SELECT id, v.variant_key, v.label, v.price_cents, true
FROM reef_products, LATERAL (VALUES
  ('4-line', '4-Line', 1800),
  ('6-line', '6-Line', 2200)
) AS v(variant_key, label, price_cents)
WHERE reef_products.slug = 'dosing-tube-organizer'
ON CONFLICT (product_id, variant_key) DO NOTHING;

INSERT INTO reef_product_variants (product_id, variant_key, label, price_cents, active)
SELECT id, v.variant_key, v.label, v.price_cents, true
FROM reef_products, LATERAL (VALUES
  ('12mm', '12mm Float Switch', 1000),
  ('16mm', '16mm Float Switch', 1000)
) AS v(variant_key, label, price_cents)
WHERE reef_products.slug = 'ato-float-switch-bracket'
ON CONFLICT (product_id, variant_key) DO NOTHING;

INSERT INTO reef_product_variants (product_id, variant_key, label, price_cents, active)
SELECT id, v.variant_key, v.label, v.price_cents, true
FROM reef_products, LATERAL (VALUES
  ('standard', '24-Hole Tray', 2200)
) AS v(variant_key, label, price_cents)
WHERE reef_products.slug = 'frag-plug-tray'
ON CONFLICT (product_id, variant_key) DO NOTHING;

INSERT INTO reef_parameter_schemas (product_id, version, schema, generator_module, generator_version, active)
SELECT id, 1, $frag$
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
      "x-unit": "mm",
      "x-label": "Glass thickness",
      "x-helpText": "Measure straight across the glass edge at the rim with calipers. Above 19mm the magnets in this design cannot hold reliably, so the range stops there.",
      "x-diagramAsset": "/reef/diagrams/glass-thickness.svg"
    },
    "tierCount": {
      "type": "integer",
      "minimum": 1,
      "maximum": 4,
      "x-label": "Tiers",
      "x-helpText": "Number of stacked frag-plug tiers. More tiers means more magnet pairs to hold the rack's weight.",
      "x-diagramAsset": "/reef/diagrams/tier-count.svg"
    },
    "widthMm": {
      "type": "number",
      "minimum": 60,
      "maximum": 250,
      "x-unit": "mm",
      "x-label": "Rack width",
      "x-helpText": "Measure the usable rim length where the rack will hang. Width also caps how many holes fit per tier.",
      "x-diagramAsset": "/reef/diagrams/rack-width.svg"
    },
    "plugHoleDiameterMm": {
      "type": "integer",
      "enum": [15, 20],
      "x-unit": "mm",
      "x-label": "Frag plug hole diameter",
      "x-helpText": "Standard frag plug stems are 15mm or 20mm. Measure your plug stem diameter, not the plug head.",
      "x-diagramAsset": "/reef/diagrams/plug-hole-diameter.svg"
    },
    "holesPerTier": {
      "type": "integer",
      "minimum": 4,
      "maximum": 12,
      "x-label": "Holes per tier",
      "x-helpText": "How many frag plugs per tier. The maximum is derived from rack width and plug hole diameter so holes never overlap.",
      "x-derivedBoundFrom": ["widthMm", "plugHoleDiameterMm"]
    },
    "color": {
      "type": "string",
      "enum": ["black", "white"],
      "default": "black",
      "x-label": "Color",
      "x-helpText": "PETG filament color. Black is the default."
    }
  }
}
$frag$::jsonb, 'frag_rack', 'v1', true
FROM reef_products WHERE slug = 'magnetic-frag-rack'
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO reef_parameter_schemas (product_id, version, schema, generator_module, generator_version, active)
SELECT id, 1, $clip$
{
  "type": "object",
  "required": ["rimThicknessMm", "rimWidthMm", "euroBrace", "meshThicknessMm", "quantity"],
  "properties": {
    "tankProfileId": {
      "type": ["string", "null"],
      "x-control": "tank-select",
      "x-label": "Tank model",
      "x-helpText": "Pick your tank model to auto-fill rim thickness and width. Not listed? Choose \"Other\" and measure by hand.",
      "x-diagramAsset": "/reef/diagrams/tank-select.svg",
      "x-autofills": ["rimThicknessMm", "rimWidthMm", "euroBrace"]
    },
    "rimThicknessMm": {
      "type": "number",
      "minimum": 3,
      "maximum": 20,
      "x-unit": "mm",
      "x-label": "Rim thickness",
      "x-helpText": "Measure the thickness of the plastic or glass rim the clip will grip, front to back.",
      "x-diagramAsset": "/reef/diagrams/rim-thickness.svg"
    },
    "rimWidthMm": {
      "type": "number",
      "minimum": 6,
      "maximum": 40,
      "x-unit": "mm",
      "x-label": "Rim width",
      "x-helpText": "Measure how far the rim extends inward from the tank's outer edge.",
      "x-diagramAsset": "/reef/diagrams/rim-width.svg"
    },
    "euroBrace": {
      "type": "boolean",
      "default": false,
      "x-label": "Euro-braced tank",
      "x-helpText": "Euro-braced tanks have a center support brace that changes how the clip needs to seat. Check your tank's top rim for a center crossbar.",
      "x-diagramAsset": "/reef/diagrams/euro-brace.svg"
    },
    "meshThicknessMm": {
      "type": "number",
      "minimum": 0.5,
      "maximum": 4,
      "x-unit": "mm",
      "x-label": "Mesh/lid thickness",
      "x-helpText": "Measure the thickness of the mesh screen or lid material the clip needs to pinch.",
      "x-diagramAsset": "/reef/diagrams/mesh-thickness.svg"
    },
    "quantity": {
      "type": "integer",
      "enum": [4, 8, 12],
      "default": 4,
      "x-label": "Quantity",
      "x-helpText": "Number of clips in the set."
    }
  }
}
$clip$::jsonb, 'lid_clip', 'v1', true
FROM reef_products WHERE slug = 'lid-mesh-clips'
ON CONFLICT (product_id, version) DO NOTHING;
