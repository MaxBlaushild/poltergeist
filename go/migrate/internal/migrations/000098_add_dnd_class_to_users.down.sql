DROP INDEX IF EXISTS idx_users_dnd_class_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_dnd_class;
ALTER TABLE users DROP COLUMN IF EXISTS dnd_class_id;