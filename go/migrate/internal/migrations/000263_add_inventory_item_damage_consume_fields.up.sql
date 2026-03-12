ALTER TABLE inventory_items
ADD COLUMN consume_deal_damage INTEGER NOT NULL DEFAULT 0,
ADD COLUMN consume_deal_damage_hits INTEGER NOT NULL DEFAULT 0,
ADD COLUMN consume_deal_damage_all_enemies INTEGER NOT NULL DEFAULT 0,
ADD COLUMN consume_deal_damage_all_enemies_hits INTEGER NOT NULL DEFAULT 0;
