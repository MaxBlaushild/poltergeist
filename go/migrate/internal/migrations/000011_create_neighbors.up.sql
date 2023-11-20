CREATE TABLE neighbors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    crystal_one_id UUID NOT NULL,
    crystal_two_id UUID NOT NULL,
    FOREIGN KEY (crystal_one_id) REFERENCES crystals(id),
    FOREIGN KEY (crystal_two_id) REFERENCES crystals(id)
);