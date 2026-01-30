CREATE TABLE albums (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL
);

CREATE INDEX idx_albums_user_id ON albums(user_id);

CREATE TABLE album_tags (
    album_id UUID NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (album_id, tag)
);

CREATE INDEX idx_album_tags_album_id ON album_tags(album_id);
CREATE INDEX idx_album_tags_tag ON album_tags(tag);
