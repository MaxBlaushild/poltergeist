ALTER TABLE reef_orders ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id);
CREATE INDEX IF NOT EXISTS reef_orders_user_id_idx ON reef_orders(user_id);
