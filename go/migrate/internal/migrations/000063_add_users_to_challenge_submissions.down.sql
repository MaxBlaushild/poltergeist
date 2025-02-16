ALTER TABLE point_of_interest_challenge_submissions
    ALTER COLUMN team_id SET NOT NULL,
    DROP CONSTRAINT fk_user,
    DROP COLUMN user_id;
