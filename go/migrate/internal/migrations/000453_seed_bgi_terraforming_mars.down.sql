DELETE FROM bgi_parameter_schemas WHERE generator_module = 'bgi_tray_set';
DELETE FROM bgi_products WHERE slug = 'terraforming-mars-tray-set';
DELETE FROM bgi_tray_templates WHERE slug = 'project_card_tray';
DELETE FROM bgi_component_manifests WHERE game_id = (SELECT id FROM bgi_games WHERE slug = 'terraforming-mars');
DELETE FROM bgi_sleeve_profiles WHERE class_key IN ('unsleeved', 'thin', 'standard', 'premium');
DELETE FROM bgi_box_profiles WHERE game_id = (SELECT id FROM bgi_games WHERE slug = 'terraforming-mars');
DELETE FROM bgi_games WHERE slug = 'terraforming-mars';
