ALTER TABLE inventory_items
  ADD COLUMN hand_item_category text,
  ADD COLUMN handedness text,
  ADD COLUMN damage_min integer,
  ADD COLUMN damage_max integer,
  ADD COLUMN swipes_per_attack integer,
  ADD COLUMN block_percentage integer,
  ADD COLUMN damage_blocked integer,
  ADD COLUMN spell_damage_bonus_percent integer;
