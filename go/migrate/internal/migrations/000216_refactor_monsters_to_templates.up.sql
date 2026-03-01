CREATE TABLE monster_templates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  base_strength INTEGER NOT NULL DEFAULT 10,
  base_dexterity INTEGER NOT NULL DEFAULT 10,
  base_constitution INTEGER NOT NULL DEFAULT 10,
  base_intelligence INTEGER NOT NULL DEFAULT 10,
  base_wisdom INTEGER NOT NULL DEFAULT 10,
  base_charisma INTEGER NOT NULL DEFAULT 10,
  legacy_monster_id UUID
);

CREATE INDEX idx_monster_templates_name ON monster_templates(name);

CREATE TABLE monster_template_spells (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  monster_template_id UUID NOT NULL REFERENCES monster_templates(id) ON DELETE CASCADE,
  spell_id UUID NOT NULL REFERENCES spells(id) ON DELETE CASCADE,
  UNIQUE(monster_template_id, spell_id)
);

CREATE INDEX idx_monster_template_spells_template_id ON monster_template_spells(monster_template_id);
CREATE INDEX idx_monster_template_spells_spell_id ON monster_template_spells(spell_id);

ALTER TABLE monsters ADD COLUMN template_id UUID;
ALTER TABLE monsters ADD COLUMN weapon_inventory_item_id INTEGER;
ALTER TABLE monsters ADD COLUMN level INTEGER NOT NULL DEFAULT 1;

ALTER TABLE monsters
  ADD CONSTRAINT fk_monsters_template_id
  FOREIGN KEY (template_id) REFERENCES monster_templates(id) ON DELETE SET NULL;

ALTER TABLE monsters
  ADD CONSTRAINT fk_monsters_weapon_inventory_item_id
  FOREIGN KEY (weapon_inventory_item_id) REFERENCES inventory_items(id) ON DELETE SET NULL;

CREATE INDEX idx_monsters_template_id ON monsters(template_id);
CREATE INDEX idx_monsters_weapon_inventory_item_id ON monsters(weapon_inventory_item_id);

INSERT INTO monster_templates (
  id,
  created_at,
  updated_at,
  name,
  description,
  base_strength,
  base_dexterity,
  base_constitution,
  base_intelligence,
  base_wisdom,
  base_charisma,
  legacy_monster_id
)
SELECT
  uuid_generate_v4(),
  NOW(),
  NOW(),
  CASE
    WHEN COALESCE(NULLIF(TRIM(name), ''), '') = '' THEN 'Monster Template'
    ELSE CONCAT(name, ' Template')
  END,
  COALESCE(description, ''),
  10,
  10,
  10,
  10,
  10,
  10,
  id
FROM monsters;

UPDATE monsters m
SET template_id = t.id
FROM monster_templates t
WHERE t.legacy_monster_id = m.id;

INSERT INTO monster_template_spells (
  id,
  created_at,
  updated_at,
  monster_template_id,
  spell_id
)
SELECT
  uuid_generate_v4(),
  NOW(),
  NOW(),
  t.id,
  msr.spell_id
FROM monster_spell_rewards msr
JOIN monster_templates t ON t.legacy_monster_id = msr.monster_id;

ALTER TABLE monster_templates DROP COLUMN legacy_monster_id;

DROP TABLE IF EXISTS monster_spell_rewards;

ALTER TABLE monsters DROP COLUMN IF EXISTS attack_damage_min;
ALTER TABLE monsters DROP COLUMN IF EXISTS attack_damage_max;
ALTER TABLE monsters DROP COLUMN IF EXISTS attack_swipes_per_attack;
