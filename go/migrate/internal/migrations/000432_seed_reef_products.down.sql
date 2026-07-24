DELETE FROM reef_parameter_schemas WHERE generator_module IN ('frag_rack', 'lid_clip');
DELETE FROM reef_product_variants WHERE product_id IN (SELECT id FROM reef_products WHERE slug IN
  ('magnetic-frag-rack', 'lid-mesh-clips', 'feeding-ring', 'dosing-tube-organizer', 'ato-float-switch-bracket', 'frag-plug-tray'));
DELETE FROM reef_products WHERE slug IN
  ('magnetic-frag-rack', 'lid-mesh-clips', 'feeding-ring', 'dosing-tube-organizer', 'ato-float-switch-bracket', 'frag-plug-tray');
