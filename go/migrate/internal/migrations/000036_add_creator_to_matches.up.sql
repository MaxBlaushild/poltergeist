ALTER TABLE matches ADD COLUMN creator_id UUID;
ALTER TABLE matches ADD CONSTRAINT fk_matches_creator FOREIGN KEY (creator_id) REFERENCES users(id);

