CREATE TABLE audit_items (
    id UUID PRIMARY KEY,
    match_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    message TEXT NOT NULL,
    FOREIGN KEY (match_id) REFERENCES matches(id)
);
