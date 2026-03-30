INSERT INTO tag_groups (id, name, created_at, updated_at, icon_url, image_url)
SELECT
  gen_random_uuid(),
  'zone_neighborhood_flavor',
  NOW(),
  NOW(),
  NULL,
  NULL
WHERE NOT EXISTS (
  SELECT 1 FROM tag_groups WHERE name = 'zone_neighborhood_flavor'
);

WITH zone_group AS (
  SELECT id
  FROM tag_groups
  WHERE name = 'zone_neighborhood_flavor'
)
INSERT INTO tags (id, tag_group_id, value, created_at, updated_at)
SELECT
  gen_random_uuid(),
  zone_group.id,
  seeded.value,
  NOW(),
  NOW()
FROM zone_group
CROSS JOIN (
  VALUES
    ('residential'),
    ('rowhouse'),
    ('brownstone'),
    ('townhouse'),
    ('tenement'),
    ('courtyard'),
    ('stoop_lined'),
    ('family_trade'),
    ('corner_store'),
    ('main_street'),
    ('high_street'),
    ('market_square'),
    ('bazaar'),
    ('shopping_arcade'),
    ('food_hall'),
    ('street_food_scene'),
    ('cafe_corridor'),
    ('tavern_row'),
    ('late_night'),
    ('lantern_lit'),
    ('nightlife'),
    ('theater_district'),
    ('music_scene'),
    ('festival_ready'),
    ('bohemian'),
    ('arts_corridor'),
    ('gallery_row'),
    ('maker_quarter'),
    ('workshop_cluster'),
    ('artisan'),
    ('foundry_belt'),
    ('warehouse_row'),
    ('dockside'),
    ('harborfront'),
    ('waterfront'),
    ('canal_edge'),
    ('bridge_approach'),
    ('railyard'),
    ('transit_crossroads'),
    ('gate_district'),
    ('old_quarter'),
    ('historic_core'),
    ('restored_block'),
    ('crumbling_edges'),
    ('affluent'),
    ('new_money'),
    ('working_class'),
    ('labor_heavy'),
    ('dockworker'),
    ('student_heavy'),
    ('campus_adjacent'),
    ('scholarly'),
    ('bureaucratic'),
    ('civic_center'),
    ('courthouse_row'),
    ('guildhall'),
    ('mercantile'),
    ('trade_route'),
    ('shopping_spine'),
    ('service_alley'),
    ('alley_maze'),
    ('backstreet'),
    ('quiet_pockets'),
    ('crowded'),
    ('bustling'),
    ('patrol_heavy'),
    ('fortified'),
    ('contested'),
    ('black_market'),
    ('clandestine'),
    ('smuggling'),
    ('undercity_edge'),
    ('haunted'),
    ('eerie'),
    ('occult'),
    ('ritual_sites'),
    ('temple_row'),
    ('shrine_lined'),
    ('memorial'),
    ('cemetery_edge'),
    ('healer_quarter'),
    ('apothecary'),
    ('garden_district'),
    ('parkland'),
    ('riverside_walk'),
    ('hilltop'),
    ('overlook'),
    ('fog_bound'),
    ('windswept'),
    ('rain_slick'),
    ('mural_filled'),
    ('community_hub'),
    ('neighborhood_pride'),
    ('pilgrim_path'),
    ('noble_presence'),
    ('mercenary_presence'),
    ('river_trade'),
    ('warehouse_nightlife'),
    ('industrial_shadow'),
    ('hidden_courts')
) AS seeded(value)
ON CONFLICT (value) DO NOTHING;

CREATE TABLE IF NOT EXISTS zone_tag_generation_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  status TEXT NOT NULL,
  context_snapshot TEXT NOT NULL DEFAULT '',
  generated_summary TEXT,
  selected_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  error_message TEXT
);

CREATE INDEX IF NOT EXISTS zone_tag_generation_jobs_zone_id_created_at_idx
  ON zone_tag_generation_jobs(zone_id, created_at DESC);

CREATE INDEX IF NOT EXISTS zone_tag_generation_jobs_created_at_idx
  ON zone_tag_generation_jobs(created_at DESC);
