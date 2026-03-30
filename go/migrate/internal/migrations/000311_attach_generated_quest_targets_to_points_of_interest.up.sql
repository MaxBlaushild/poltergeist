ALTER TABLE monster_encounters
  ADD COLUMN IF NOT EXISTS point_of_interest_id UUID REFERENCES points_of_interest(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS monster_encounters_point_of_interest_id_idx
  ON monster_encounters(point_of_interest_id);

WITH poi_candidates AS (
  SELECT
    poi.id,
    pz.zone_id,
    CAST(TRIM(poi.lat) AS double precision) AS latitude,
    CAST(TRIM(poi.lng) AS double precision) AS longitude
  FROM points_of_interest AS poi
  JOIN point_of_interest_zones AS pz
    ON pz.point_of_interest_id = poi.id
   AND pz.deleted_at IS NULL
  WHERE TRIM(COALESCE(poi.lat, '')) ~ '^-?[0-9]+([.][0-9]+)?$'
    AND TRIM(COALESCE(poi.lng, '')) ~ '^-?[0-9]+([.][0-9]+)?$'
)
UPDATE challenges AS ch
SET point_of_interest_id = poi.id
FROM poi_candidates AS poi
JOIN quest_nodes AS qn ON qn.challenge_id = ch.id
WHERE ch.point_of_interest_id IS NULL
  AND ch.zone_id = poi.zone_id
  AND ABS(ch.latitude - poi.latitude) < 0.0000001
  AND ABS(ch.longitude - poi.longitude) < 0.0000001;

WITH poi_candidates AS (
  SELECT
    poi.id,
    pz.zone_id,
    CAST(TRIM(poi.lat) AS double precision) AS latitude,
    CAST(TRIM(poi.lng) AS double precision) AS longitude
  FROM points_of_interest AS poi
  JOIN point_of_interest_zones AS pz
    ON pz.point_of_interest_id = poi.id
   AND pz.deleted_at IS NULL
  WHERE TRIM(COALESCE(poi.lat, '')) ~ '^-?[0-9]+([.][0-9]+)?$'
    AND TRIM(COALESCE(poi.lng, '')) ~ '^-?[0-9]+([.][0-9]+)?$'
)
UPDATE scenarios AS sc
SET point_of_interest_id = poi.id
FROM poi_candidates AS poi
JOIN quest_nodes AS qn ON qn.scenario_id = sc.id
WHERE sc.point_of_interest_id IS NULL
  AND sc.zone_id = poi.zone_id
  AND ABS(sc.latitude - poi.latitude) < 0.0000001
  AND ABS(sc.longitude - poi.longitude) < 0.0000001;

WITH poi_candidates AS (
  SELECT
    poi.id,
    pz.zone_id,
    CAST(TRIM(poi.lat) AS double precision) AS latitude,
    CAST(TRIM(poi.lng) AS double precision) AS longitude
  FROM points_of_interest AS poi
  JOIN point_of_interest_zones AS pz
    ON pz.point_of_interest_id = poi.id
   AND pz.deleted_at IS NULL
  WHERE TRIM(COALESCE(poi.lat, '')) ~ '^-?[0-9]+([.][0-9]+)?$'
    AND TRIM(COALESCE(poi.lng, '')) ~ '^-?[0-9]+([.][0-9]+)?$'
)
UPDATE monster_encounters AS me
SET point_of_interest_id = poi.id
FROM poi_candidates AS poi
JOIN quest_nodes AS qn ON qn.monster_encounter_id = me.id
WHERE me.point_of_interest_id IS NULL
  AND me.zone_id = poi.zone_id
  AND ABS(me.latitude - poi.latitude) < 0.0000001
  AND ABS(me.longitude - poi.longitude) < 0.0000001;
