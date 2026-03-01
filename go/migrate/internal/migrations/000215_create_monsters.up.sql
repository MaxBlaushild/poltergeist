CREATE TABLE monsters (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  image_url TEXT NOT NULL DEFAULT '',
  thumbnail_url TEXT NOT NULL DEFAULT '',
  zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  geometry geometry(Point,4326) NOT NULL,
  attack_damage_min INTEGER NOT NULL DEFAULT 1,
  attack_damage_max INTEGER NOT NULL DEFAULT 1,
  attack_swipes_per_attack INTEGER NOT NULL DEFAULT 1,
  reward_experience INTEGER NOT NULL DEFAULT 0,
  reward_gold INTEGER NOT NULL DEFAULT 0,
  image_generation_status TEXT NOT NULL DEFAULT 'none',
  image_generation_error TEXT
);

CREATE INDEX idx_monsters_zone_id ON monsters(zone_id);
CREATE INDEX idx_monsters_geometry ON monsters USING GIST(geometry);
CREATE INDEX idx_monsters_name ON monsters(name);

CREATE TABLE monster_spell_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  monster_id UUID NOT NULL REFERENCES monsters(id) ON DELETE CASCADE,
  spell_id UUID NOT NULL REFERENCES spells(id) ON DELETE CASCADE,
  UNIQUE(monster_id, spell_id)
);

CREATE INDEX idx_monster_spell_rewards_monster_id ON monster_spell_rewards(monster_id);
CREATE INDEX idx_monster_spell_rewards_spell_id ON monster_spell_rewards(spell_id);

CREATE TABLE monster_item_rewards (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  monster_id UUID NOT NULL REFERENCES monsters(id) ON DELETE CASCADE,
  inventory_item_id INTEGER NOT NULL REFERENCES inventory_items(id),
  quantity INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_monster_item_rewards_monster_id ON monster_item_rewards(monster_id);
CREATE INDEX idx_monster_item_rewards_inventory_item_id ON monster_item_rewards(inventory_item_id);
