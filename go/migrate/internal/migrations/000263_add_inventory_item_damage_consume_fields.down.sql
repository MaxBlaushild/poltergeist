ALTER TABLE inventory_items
DROP COLUMN IF EXISTS consume_deal_damage,
DROP COLUMN IF EXISTS consume_deal_damage_hits,
DROP COLUMN IF EXISTS consume_deal_damage_all_enemies,
DROP COLUMN IF EXISTS consume_deal_damage_all_enemies_hits;
