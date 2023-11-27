CREATE TABLE sonar_surveys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    referrer_id UUID NOT NULL,
    progenitor_id UUID,
    FOREIGN KEY (referrer_id) REFERENCES users(id),
    FOREIGN KEY (progenitor_id) REFERENCES sonar_surveys(id)
);