ALTER TABLE sonar_surveys ADD COLUMN user_id UUID NOT NULL;
ALTER TABLE sonar_surveys ADD CONSTRAINT fk_user_surveys FOREIGN KEY (user_id) REFERENCES users(id);