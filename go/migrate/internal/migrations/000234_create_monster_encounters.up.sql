CREATE TABLE monster_encounters (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  image_url TEXT NOT NULL DEFAULT '',
  thumbnail_url TEXT NOT NULL DEFAULT '',
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  geometry geometry(Point,4326)
);

CREATE INDEX idx_monster_encounters_zone_id ON monster_encounters(zone_id);
CREATE INDEX idx_monster_encounters_geometry ON monster_encounters USING GIST(geometry);

CREATE TABLE monster_encounter_members (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  monster_encounter_id UUID NOT NULL REFERENCES monster_encounters(id) ON DELETE CASCADE,
  monster_id UUID NOT NULL REFERENCES monsters(id) ON DELETE CASCADE,
  slot INTEGER NOT NULL CHECK (slot >= 1 AND slot <= 9)
);

CREATE UNIQUE INDEX monster_encounter_members_encounter_monster_uq
  ON monster_encounter_members(monster_encounter_id, monster_id);
CREATE UNIQUE INDEX monster_encounter_members_encounter_slot_uq
  ON monster_encounter_members(monster_encounter_id, slot);
CREATE INDEX idx_monster_encounter_members_monster_id
  ON monster_encounter_members(monster_id);

ALTER TABLE quest_nodes
  ADD COLUMN monster_encounter_id UUID REFERENCES monster_encounters(id) ON DELETE SET NULL;

CREATE INDEX quest_nodes_monster_encounter_idx ON quest_nodes(monster_encounter_id);

CREATE TEMP TABLE tmp_monster_encounter_backfill (
  monster_id UUID PRIMARY KEY,
  encounter_id UUID NOT NULL
);

INSERT INTO tmp_monster_encounter_backfill (monster_id, encounter_id)
SELECT id, uuid_generate_v4()
FROM monsters;

INSERT INTO monster_encounters (
  id,
  created_at,
  updated_at,
  name,
  description,
  image_url,
  thumbnail_url,
  zone_id,
  latitude,
  longitude,
  geometry
)
SELECT
  tmp.encounter_id,
  NOW(),
  NOW(),
  CASE
    WHEN NULLIF(BTRIM(monsters.name), '') IS NOT NULL THEN FORMAT('%s Encounter', NULLIF(BTRIM(monsters.name), ''))
    ELSE 'Monster Encounter'
  END,
  COALESCE(NULLIF(BTRIM(monsters.description), ''), ''),
  COALESCE(NULLIF(BTRIM(monsters.image_url), ''), ''),
  COALESCE(NULLIF(BTRIM(monsters.thumbnail_url), ''), COALESCE(NULLIF(BTRIM(monsters.image_url), ''), '')),
  monsters.zone_id,
  monsters.latitude,
  monsters.longitude,
  monsters.geometry
FROM tmp_monster_encounter_backfill tmp
JOIN monsters ON monsters.id = tmp.monster_id;

INSERT INTO monster_encounter_members (
  id,
  created_at,
  updated_at,
  monster_encounter_id,
  monster_id,
  slot
)
SELECT
  uuid_generate_v4(),
  NOW(),
  NOW(),
  tmp.encounter_id,
  tmp.monster_id,
  1
FROM tmp_monster_encounter_backfill tmp;

UPDATE quest_nodes qn
SET monster_encounter_id = tmp.encounter_id
FROM tmp_monster_encounter_backfill tmp
WHERE qn.monster_id = tmp.monster_id
  AND qn.monster_encounter_id IS NULL;

DROP TABLE tmp_monster_encounter_backfill;

ALTER TABLE quest_nodes
  DROP CONSTRAINT IF EXISTS quest_nodes_exactly_one_target;

ALTER TABLE quest_nodes
  ADD CONSTRAINT quest_nodes_exactly_one_target
  CHECK (
    (
      (CASE WHEN scenario_id IS NOT NULL THEN 1 ELSE 0 END) +
      (CASE WHEN challenge_id IS NOT NULL THEN 1 ELSE 0 END) +
      (CASE WHEN monster_id IS NOT NULL OR monster_encounter_id IS NOT NULL THEN 1 ELSE 0 END)
    ) = 1
  );
