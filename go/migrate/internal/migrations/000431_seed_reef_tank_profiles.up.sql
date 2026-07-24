-- R-3.4: seed the tank-profile research backlog. This environment has no
-- outbound web access available to this migration's author (every fetch to a
-- manufacturer spec page returned 403), so none of these rows can honestly be
-- marked verified=true — that would require a human to actually visit each
-- manufacturer's spec page, confirm rim/glass thickness, and record source_url.
-- Per R-3.4 these rows are seeded unverified on purpose: they track the
-- research backlog and are excluded from configurator dropdowns
-- (WHERE verified = true) until an operator fills in source_url and flips the
-- flag. Placeholder dimensions are order-of-magnitude typical values for each
-- line, NOT to be trusted for manufacturing until verified.
INSERT INTO reef_tank_profiles
  (manufacturer, model, rim_thickness_mm, rim_width_mm, glass_thickness_mm, euro_brace, internal_dims, verified, source_url)
VALUES
  ('Waterbox', 'Reef 54.3', 8, 25, 8, true, '{}'::jsonb, false, ''),
  ('Waterbox', 'Reef 90.4', 8, 25, 10, true, '{}'::jsonb, false, ''),
  ('Waterbox', 'Reef 130.4', 8, 25, 10, true, '{}'::jsonb, false, ''),
  ('Waterbox', 'Clear 20', 6, 20, 8, false, '{}'::jsonb, false, ''),
  ('Red Sea', 'Reefer 170 G2', 8, 25, 8, true, '{}'::jsonb, false, ''),
  ('Red Sea', 'Reefer 250 G2', 8, 25, 10, true, '{}'::jsonb, false, ''),
  ('Red Sea', 'Reefer 350 G2', 8, 25, 10, true, '{}'::jsonb, false, ''),
  ('Innovative Marine', 'Fusion 20', 6, 20, 8, false, '{}'::jsonb, false, ''),
  ('Innovative Marine', 'NUVO 20', 6, 20, 8, false, '{}'::jsonb, false, ''),
  ('Cade', 'Reef Pro 1200', 8, 25, 10, true, '{}'::jsonb, false, ''),
  ('Fiji Cube', 'Frag 25', 6, 20, 8, false, '{}'::jsonb, false, '')
ON CONFLICT (manufacturer, model) DO NOTHING;
