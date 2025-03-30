CREATE TABLE IF NOT EXISTS tag_entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    point_of_interest_id UUID REFERENCES points_of_interest(id) ON DELETE CASCADE,
    point_of_interest_group_id UUID REFERENCES point_of_interest_groups(id) ON DELETE CASCADE,
    point_of_interest_challenge_id UUID REFERENCES point_of_interest_challenges(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tag_entities_poi_id ON tag_entities(point_of_interest_id);
CREATE INDEX IF NOT EXISTS idx_tag_entities_poi_group_id ON tag_entities(point_of_interest_group_id);
CREATE INDEX IF NOT EXISTS idx_tag_entities_poi_challenge_id ON tag_entities(point_of_interest_challenge_id);
