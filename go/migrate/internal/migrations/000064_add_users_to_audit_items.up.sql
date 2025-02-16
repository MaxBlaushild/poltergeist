ALTER TABLE audit_items
    ALTER COLUMN team_id DROP NOT NULL,
    ADD COLUMN user_id UUID,
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id);
