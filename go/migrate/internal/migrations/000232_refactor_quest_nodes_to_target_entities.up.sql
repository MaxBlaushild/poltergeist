WITH legacy_poi_nodes AS (
  SELECT
    qn.id AS node_id,
    uuid_generate_v4() AS new_challenge_id,
    COALESCE(q.zone_id, pz.zone_id, fallback_zone.id) AS zone_id,
    CASE
      WHEN poi.lat ~ '^-?[0-9]+(\.[0-9]+)?$' THEN poi.lat::double precision
      ELSE 0
    END AS latitude,
    CASE
      WHEN poi.lng ~ '^-?[0-9]+(\.[0-9]+)?$' THEN poi.lng::double precision
      ELSE 0
    END AS longitude,
    COALESCE(
      NULLIF(BTRIM(qnc.question), ''),
      FORMAT('Roleplay an action at %s and provide proof.', COALESCE(NULLIF(BTRIM(poi.name), ''), 'this location'))
    ) AS question,
    COALESCE(
      NULLIF(BTRIM(q.description), ''),
      NULLIF(BTRIM(poi.description), ''),
      ''
    ) AS description,
    COALESCE(qnc.reward, 0) AS reward,
    qnc.inventory_item_id AS inventory_item_id,
    COALESCE(NULLIF(BTRIM(qn.submission_type), ''), 'photo') AS submission_type,
    COALESCE(qnc.difficulty, 0) AS difficulty,
    COALESCE(qnc.stat_tags, '[]'::jsonb) AS stat_tags,
    qnc.proficiency AS proficiency
  FROM quest_nodes qn
  JOIN quests q ON q.id = qn.quest_id
  JOIN points_of_interest poi ON poi.id = qn.point_of_interest_id
  LEFT JOIN point_of_interest_zones pz ON pz.point_of_interest_id = poi.id
  LEFT JOIN LATERAL (
    SELECT id
    FROM zones
    ORDER BY created_at ASC
    LIMIT 1
  ) fallback_zone ON TRUE
  LEFT JOIN LATERAL (
    SELECT
      question,
      reward,
      inventory_item_id,
      difficulty,
      stat_tags,
      proficiency
    FROM quest_node_challenges
    WHERE quest_node_id = qn.id
    ORDER BY tier ASC, created_at ASC
    LIMIT 1
  ) qnc ON TRUE
  WHERE qn.challenge_id IS NULL
    AND qn.scenario_id IS NULL
    AND qn.monster_id IS NULL
    AND qn.point_of_interest_id IS NOT NULL
    AND COALESCE(q.zone_id, pz.zone_id, fallback_zone.id) IS NOT NULL
),
inserted_legacy_poi AS (
  INSERT INTO challenges (
    id,
    created_at,
    updated_at,
    zone_id,
    latitude,
    longitude,
    geometry,
    question,
    description,
    reward,
    inventory_item_id,
    submission_type,
    difficulty,
    stat_tags,
    proficiency,
    image_url,
    thumbnail_url
  )
  SELECT
    lpn.new_challenge_id,
    NOW(),
    NOW(),
    lpn.zone_id,
    lpn.latitude,
    lpn.longitude,
    ST_SetSRID(ST_MakePoint(lpn.longitude, lpn.latitude), 4326),
    lpn.question,
    lpn.description,
    lpn.reward,
    lpn.inventory_item_id,
    lpn.submission_type,
    lpn.difficulty,
    lpn.stat_tags,
    lpn.proficiency,
    '',
    ''
  FROM legacy_poi_nodes lpn
)
UPDATE quest_nodes qn
SET challenge_id = lpn.new_challenge_id
FROM legacy_poi_nodes lpn
WHERE qn.id = lpn.node_id;

WITH legacy_polygon_nodes AS (
  SELECT
    qn.id AS node_id,
    uuid_generate_v4() AS new_challenge_id,
    COALESCE(q.zone_id, fallback_zone.id) AS zone_id,
    ST_Y(ST_PointOnSurface(qn.polygon)) AS latitude,
    ST_X(ST_PointOnSurface(qn.polygon)) AS longitude,
    COALESCE(
      NULLIF(BTRIM(qnc.question), ''),
      'Roleplay an action in this area and provide proof.'
    ) AS question,
    COALESCE(NULLIF(BTRIM(q.description), ''), '') AS description,
    COALESCE(qnc.reward, 0) AS reward,
    qnc.inventory_item_id AS inventory_item_id,
    COALESCE(NULLIF(BTRIM(qn.submission_type), ''), 'photo') AS submission_type,
    COALESCE(qnc.difficulty, 0) AS difficulty,
    COALESCE(qnc.stat_tags, '[]'::jsonb) AS stat_tags,
    qnc.proficiency AS proficiency
  FROM quest_nodes qn
  JOIN quests q ON q.id = qn.quest_id
  LEFT JOIN LATERAL (
    SELECT id
    FROM zones
    ORDER BY created_at ASC
    LIMIT 1
  ) fallback_zone ON TRUE
  LEFT JOIN LATERAL (
    SELECT
      question,
      reward,
      inventory_item_id,
      difficulty,
      stat_tags,
      proficiency
    FROM quest_node_challenges
    WHERE quest_node_id = qn.id
    ORDER BY tier ASC, created_at ASC
    LIMIT 1
  ) qnc ON TRUE
  WHERE qn.challenge_id IS NULL
    AND qn.scenario_id IS NULL
    AND qn.monster_id IS NULL
    AND qn.point_of_interest_id IS NULL
    AND qn.polygon IS NOT NULL
    AND COALESCE(q.zone_id, fallback_zone.id) IS NOT NULL
),
inserted_legacy_polygon AS (
  INSERT INTO challenges (
    id,
    created_at,
    updated_at,
    zone_id,
    latitude,
    longitude,
    geometry,
    question,
    description,
    reward,
    inventory_item_id,
    submission_type,
    difficulty,
    stat_tags,
    proficiency,
    image_url,
    thumbnail_url
  )
  SELECT
    lgn.new_challenge_id,
    NOW(),
    NOW(),
    lgn.zone_id,
    lgn.latitude,
    lgn.longitude,
    ST_SetSRID(ST_MakePoint(lgn.longitude, lgn.latitude), 4326),
    lgn.question,
    lgn.description,
    lgn.reward,
    lgn.inventory_item_id,
    lgn.submission_type,
    lgn.difficulty,
    lgn.stat_tags,
    lgn.proficiency,
    '',
    ''
  FROM legacy_polygon_nodes lgn
)
UPDATE quest_nodes qn
SET challenge_id = lgn.new_challenge_id
FROM legacy_polygon_nodes lgn
WHERE qn.id = lgn.node_id;

WITH unresolved_nodes AS (
  SELECT
    qn.id AS node_id,
    uuid_generate_v4() AS new_challenge_id,
    COALESCE(q.zone_id, fallback_zone.id) AS zone_id,
    COALESCE(
      CASE
        WHEN zone_poi.lat ~ '^-?[0-9]+(\.[0-9]+)?$' THEN zone_poi.lat::double precision
        ELSE NULL
      END,
      ST_Y(ST_PointOnSurface(zone_record.boundary)),
      0
    ) AS latitude,
    COALESCE(
      CASE
        WHEN zone_poi.lng ~ '^-?[0-9]+(\.[0-9]+)?$' THEN zone_poi.lng::double precision
        ELSE NULL
      END,
      ST_X(ST_PointOnSurface(zone_record.boundary)),
      0
    ) AS longitude,
    COALESCE(
      NULLIF(BTRIM(qnc.question), ''),
      'Roleplay an action somewhere in this zone and provide proof.'
    ) AS question,
    COALESCE(NULLIF(BTRIM(q.description), ''), '') AS description,
    COALESCE(qnc.reward, 0) AS reward,
    qnc.inventory_item_id AS inventory_item_id,
    COALESCE(NULLIF(BTRIM(qn.submission_type), ''), 'photo') AS submission_type,
    COALESCE(qnc.difficulty, 0) AS difficulty,
    COALESCE(qnc.stat_tags, '[]'::jsonb) AS stat_tags,
    qnc.proficiency AS proficiency
  FROM quest_nodes qn
  JOIN quests q ON q.id = qn.quest_id
  LEFT JOIN zones zone_record ON zone_record.id = q.zone_id
  LEFT JOIN LATERAL (
    SELECT id
    FROM zones
    ORDER BY created_at ASC
    LIMIT 1
  ) fallback_zone ON TRUE
  LEFT JOIN LATERAL (
    SELECT
      poi.lat,
      poi.lng
    FROM point_of_interest_zones pz
    JOIN points_of_interest poi ON poi.id = pz.point_of_interest_id
    WHERE pz.zone_id = COALESCE(q.zone_id, fallback_zone.id)
    ORDER BY poi.created_at ASC
    LIMIT 1
  ) zone_poi ON TRUE
  LEFT JOIN LATERAL (
    SELECT
      question,
      reward,
      inventory_item_id,
      difficulty,
      stat_tags,
      proficiency
    FROM quest_node_challenges
    WHERE quest_node_id = qn.id
    ORDER BY tier ASC, created_at ASC
    LIMIT 1
  ) qnc ON TRUE
  WHERE qn.challenge_id IS NULL
    AND qn.scenario_id IS NULL
    AND qn.monster_id IS NULL
    AND qn.point_of_interest_id IS NULL
    AND qn.polygon IS NULL
    AND COALESCE(q.zone_id, fallback_zone.id) IS NOT NULL
),
inserted_unresolved AS (
  INSERT INTO challenges (
    id,
    created_at,
    updated_at,
    zone_id,
    latitude,
    longitude,
    geometry,
    question,
    description,
    reward,
    inventory_item_id,
    submission_type,
    difficulty,
    stat_tags,
    proficiency,
    image_url,
    thumbnail_url
  )
  SELECT
    un.new_challenge_id,
    NOW(),
    NOW(),
    un.zone_id,
    un.latitude,
    un.longitude,
    ST_SetSRID(ST_MakePoint(un.longitude, un.latitude), 4326),
    un.question,
    un.description,
    un.reward,
    un.inventory_item_id,
    un.submission_type,
    un.difficulty,
    un.stat_tags,
    un.proficiency,
    '',
    ''
  FROM unresolved_nodes un
)
UPDATE quest_nodes qn
SET challenge_id = un.new_challenge_id
FROM unresolved_nodes un
WHERE qn.id = un.node_id;

DROP INDEX IF EXISTS quest_nodes_poi_idx;

ALTER TABLE quest_nodes
  DROP COLUMN IF EXISTS point_of_interest_id,
  DROP COLUMN IF EXISTS polygon;

ALTER TABLE quest_nodes
  DROP CONSTRAINT IF EXISTS quest_nodes_exactly_one_target;

ALTER TABLE quest_nodes
  ADD CONSTRAINT quest_nodes_exactly_one_target
  CHECK (num_nonnulls(scenario_id, monster_id, challenge_id) = 1) NOT VALID;
