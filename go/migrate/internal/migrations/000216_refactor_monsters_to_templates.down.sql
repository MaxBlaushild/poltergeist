ALTER TABLE monsters ADD COLUMN attack_damage_min INTEGER NOT NULL DEFAULT 1;
ALTER TABLE monsters ADD COLUMN attack_damage_max INTEGER NOT NULL DEFAULT 1;
ALTER TABLE monsters ADD COLUMN attack_swipes_per_attack INTEGER NOT NULL DEFAULT 1;

UPDATE monsters m
SET
  attack_damage_min = COALESCE(i.damage_min, 1),
  attack_damage_max = COALESCE(i.damage_max, 1),
  attack_swipes_per_attack = COALESCE(i.swipes_per_attack, 1)
FROM inventory_items i
WHERE m.weapon_inventory_item_id = i.id;

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

INSERT INTO monster_spell_rewards (
  id,
  created_at,
  updated_at,
  monster_id,
  spell_id
)
SELECT
  uuid_generate_v4(),
  NOW(),
  NOW(),
  m.id,
  mts.spell_id
FROM monsters m
JOIN monster_template_spells mts ON mts.monster_template_id = m.template_id
ON CONFLICT (monster_id, spell_id) DO NOTHING;

DROP INDEX IF EXISTS idx_monsters_template_id;
DROP INDEX IF EXISTS idx_monsters_weapon_inventory_item_id;

ALTER TABLE monsters DROP CONSTRAINT IF EXISTS fk_monsters_template_id;
ALTER TABLE monsters DROP CONSTRAINT IF EXISTS fk_monsters_weapon_inventory_item_id;

ALTER TABLE monsters DROP COLUMN IF EXISTS template_id;
ALTER TABLE monsters DROP COLUMN IF EXISTS weapon_inventory_item_id;
ALTER TABLE monsters DROP COLUMN IF EXISTS level;

DROP TABLE IF EXISTS monster_template_spells;
DROP TABLE IF EXISTS monster_templates;
