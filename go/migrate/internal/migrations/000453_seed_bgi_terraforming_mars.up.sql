-- Seeds the first bgi launch game end to end (R-1.2's "seed one launch game
-- fully before adding others"). Every measurement here is sourced from
-- published/community references, not a physical measurement of the user's
-- own copy — hence verified = false throughout on every box/sleeve/manifest
-- row. That is the actual trust signal (mirroring reef_tank_profiles'
-- verified column: it excludes a row from convenience surfaces, it does not
-- gate the whole product's liveness) — a "pending physical verification"
-- notice must render wherever this data is used (product page, fit
-- indicator) per the frontend's own requirement. bgi_games.active and
-- bgi_products.active are TRUE so the configurator and checkout are fully
-- functional and testable end to end (acceptance criteria 1/2/4/6/8/9 all
-- require actually running them) — going live for real public traffic and
-- real orders is a separate, later decision made by not deploying/
-- announcing trays.forteus.tech until the user's own hand-verification
-- (acceptance criterion 3) is done, not by a DB flag blocking checkout.
-- Sources:
--
-- - Box interior ~286x286mm: a Thingiverse insert designer's stated
--   interior dimensions for this box. Depth (72.5mm) has no independent
--   source found in this pass — depth_is_placeholder=true flags it as
--   materially lower-confidence than length/width, not a real measurement.
-- - 12 unique corporation cards: Stronghold Games' official Kickstarter FAQ
--   page states this explicitly (higher confidence than the project-card
--   count) — still seeded unverified, since verification here is a
--   physical-check gate, not a research-confidence one.
-- - 208 project cards: commonly and consistently cited across community
--   sources; not confirmed against the official rulebook.
-- - Sleeve thickness classes: derived from published board-game sleeve
--   market figures (unsleeved card ~0.30mm; sleeve film commonly
--   45-160 microns depending on class, ~70-100 microns cited as the
--   community "sweet spot" for board-game sleeving).

INSERT INTO bgi_games (slug, name, publisher, year_published, active) VALUES
  ('terraforming-mars', 'Terraforming Mars', 'FryxGames', 2016, true)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO bgi_box_profiles (game_id, slug, label, source, interior_length_mm, interior_width_mm, interior_depth_mm, depth_is_placeholder, verified, source_url, measurement_notes)
SELECT id, 'original', 'Original Box', 'original', 286, 286, 72.5, true, false, '',
  'Interior L/W (286x286mm) from a Thingiverse insert designer''s stated dimensions for this box. Depth (72.5mm) is an unsourced placeholder pending physical measurement — do not treat as measured.'
FROM bgi_games WHERE slug = 'terraforming-mars'
ON CONFLICT (game_id, slug) DO NOTHING;

INSERT INTO bgi_sleeve_profiles (class_key, label, base_card_thickness_mm, sleeve_material_thickness_mm, verified, source_url) VALUES
  ('unsleeved', 'Unsleeved', 0.30, 0.0, false, ''),
  ('thin', 'Thin / Penny Sleeves', 0.30, 0.045, false, ''),
  ('standard', 'Standard Sleeves', 0.30, 0.085, false, ''),
  ('premium', 'Premium / Thick Sleeves', 0.30, 0.16, false, '')
ON CONFLICT (class_key) DO NOTHING;

INSERT INTO bgi_component_manifests (game_id, expansion_id, component_type, card_width_mm, card_height_mm, count, verified, source_url, notes)
SELECT id, NULL, 'project_card', 44, 68, 208, false, '',
  'Commonly cited project-card count across community sources; not confirmed against the official rulebook. Card dims are a standard "mini/Euro" card size estimate, not independently measured.'
FROM bgi_games WHERE slug = 'terraforming-mars'
ON CONFLICT (game_id, expansion_id, component_type) DO NOTHING;

INSERT INTO bgi_component_manifests (game_id, expansion_id, component_type, card_width_mm, card_height_mm, count, verified, source_url, notes)
SELECT id, NULL, 'corporation_card', 44, 68, 12, false, '',
  '12 unique corporation cards is stated on Stronghold Games'' official Kickstarter FAQ page — higher research confidence than the project-card count, but still unverified pending the physical-check gate. No tray template exists for this component type in v1 (see bgi_tray_templates) — the set-assembly service surfaces it as an unassembled component rather than silently dropping it.'
FROM bgi_games WHERE slug = 'terraforming-mars'
ON CONFLICT (game_id, expansion_id, component_type) DO NOTHING;

INSERT INTO bgi_tray_templates (slug, generator_module, generator_version, component_type, description, active) VALUES
  ('project_card_tray', 'bgi_card_tray', 'v1', 'project_card',
   'An open-top card well with a finger scoop, sized to a customer-configured card count and sleeve thickness by the set-assembly service.', true)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO bgi_products (game_id, slug, name, kind, description, material, base_price_cents, images, active)
SELECT id, 'terraforming-mars-tray-set', 'Terraforming Mars Tray Set', 'configurable',
  'A made-to-order freestanding tray set for Terraforming Mars, sized to your sleeve class and target box. Unaffiliated with and unendorsed by FryxGames — see /games/terraforming-mars/compatibility.',
  'PETG', 0, '[]'::jsonb, true
FROM bgi_games WHERE slug = 'terraforming-mars'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO bgi_parameter_schemas (product_id, version, schema, generator_module, generator_version, active)
SELECT id, 1, $bgi$
{
  "type": "object",
  "required": ["sleeveProfileId", "boxProfileId", "color"],
  "properties": {
    "sleeveProfileId": {
      "type": "string",
      "x-control": "sleeve-select",
      "x-label": "Sleeve profile",
      "x-helpText": "Pick the sleeve class your cards use (or unsleeved). This directly changes card-well depth — the single parameter that matters most for a good fit."
    },
    "boxProfileId": {
      "type": "string",
      "x-control": "box-select",
      "x-label": "Box target",
      "x-helpText": "Pick which box the trays need to fit — the original game box, or a listed aftermarket box."
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
$bgi$::jsonb, 'bgi_tray_set', 'v1', true
FROM bgi_products WHERE slug = 'terraforming-mars-tray-set'
ON CONFLICT (product_id, version) DO NOTHING;
