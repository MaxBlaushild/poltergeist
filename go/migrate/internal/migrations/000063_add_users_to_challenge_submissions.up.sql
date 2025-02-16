ALTER TABLE point_of_interest_challenge_submissions
    ALTER COLUMN team_id DROP NOT NULL,
    ADD COLUMN user_id UUID,
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id);
