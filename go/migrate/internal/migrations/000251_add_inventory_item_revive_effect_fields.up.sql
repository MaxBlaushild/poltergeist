ALTER TABLE inventory_items
ADD COLUMN IF NOT EXISTS consume_revive_party_member_health INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS consume_revive_all_downed_party_members_health INTEGER NOT NULL DEFAULT 0;
