DROP INDEX IF EXISTS reef_orders_user_id_idx;
ALTER TABLE reef_orders DROP COLUMN IF EXISTS user_id;
