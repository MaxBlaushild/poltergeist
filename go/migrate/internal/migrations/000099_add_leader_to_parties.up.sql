ALTER TABLE parties
ADD COLUMN leader_id UUID REFERENCES users(id);
