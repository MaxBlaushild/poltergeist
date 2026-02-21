ALTER TABLE inventory_items
  ADD COLUMN equip_slot text,
  ADD COLUMN strength_mod integer NOT NULL DEFAULT 0,
  ADD COLUMN dexterity_mod integer NOT NULL DEFAULT 0,
  ADD COLUMN constitution_mod integer NOT NULL DEFAULT 0,
  ADD COLUMN intelligence_mod integer NOT NULL DEFAULT 0,
  ADD COLUMN wisdom_mod integer NOT NULL DEFAULT 0,
  ADD COLUMN charisma_mod integer NOT NULL DEFAULT 0;

CREATE TABLE user_equipment (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  slot text NOT NULL,
  owned_inventory_item_id uuid NOT NULL REFERENCES owned_inventory_items(id) ON DELETE CASCADE,
  created_at timestamp NOT NULL DEFAULT now(),
  updated_at timestamp NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX user_equipment_user_slot_idx ON user_equipment(user_id, slot);
CREATE UNIQUE INDEX user_equipment_owned_item_idx ON user_equipment(owned_inventory_item_id);
CREATE INDEX user_equipment_user_idx ON user_equipment(user_id);
