ALTER TABLE audit_items
    ALTER COLUMN team_id SET NOT NULL,
    DROP CONSTRAINT fk_user,
    DROP COLUMN user_id;
