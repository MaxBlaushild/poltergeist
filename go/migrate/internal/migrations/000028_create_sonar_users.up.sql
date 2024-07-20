CREATE TABLE sonar_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    viewer_id UUID NOT NULL,
    viewee_id UUID NOT NULL,
    FOREIGN KEY (viewer_id) REFERENCES users(id),
    FOREIGN KEY (viewee_id) REFERENCES users(id)
);
