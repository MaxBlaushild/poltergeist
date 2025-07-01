ALTER TABLE users ADD COLUMN dnd_class_id UUID;
ALTER TABLE users ADD CONSTRAINT fk_users_dnd_class FOREIGN KEY (dnd_class_id) REFERENCES dnd_classes(id);
CREATE INDEX idx_users_dnd_class_id ON users(dnd_class_id);