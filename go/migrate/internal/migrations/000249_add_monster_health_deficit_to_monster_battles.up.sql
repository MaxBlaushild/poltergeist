ALTER TABLE monster_battles
ADD COLUMN IF NOT EXISTS monster_health_deficit INTEGER NOT NULL DEFAULT 0;
