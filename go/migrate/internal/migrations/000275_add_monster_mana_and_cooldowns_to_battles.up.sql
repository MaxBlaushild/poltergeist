ALTER TABLE monster_battles
ADD COLUMN monster_mana_deficit INTEGER NOT NULL DEFAULT 0,
ADD COLUMN monster_ability_cooldowns JSONB NOT NULL DEFAULT '{}'::jsonb;
