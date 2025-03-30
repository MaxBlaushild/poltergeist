CREATE TABLE IF NOT EXISTS tag_groups (
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tag_groups_deleted_at ON tag_groups(deleted_at);

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY,
    tag_group_id UUID NOT NULL REFERENCES tag_groups(id),
    value VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tags_deleted_at ON tags(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tags_tag_group_id ON tags(tag_group_id);

