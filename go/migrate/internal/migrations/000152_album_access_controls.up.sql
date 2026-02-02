-- Album members: admin or poster role. Owner (user_id on albums) is implicit admin.
CREATE TABLE album_members (
    album_id UUID NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('admin', 'poster')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (album_id, user_id)
);
CREATE INDEX idx_album_members_user_id ON album_members(user_id);

-- Album invites (role = role to assign when accepted)
CREATE TABLE album_invites (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    album_id UUID NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    invited_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invited_by_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'poster' CHECK (role IN ('admin', 'poster')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(album_id, invited_user_id)
);
CREATE INDEX idx_album_invites_album_id ON album_invites(album_id);
CREATE INDEX idx_album_invites_invited_user_id ON album_invites(invited_user_id);

-- Explicit album posts (add/remove). Replaces tag-based for post list when non-empty.
CREATE TABLE album_posts (
    album_id UUID NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (album_id, post_id)
);
CREATE INDEX idx_album_posts_album_id ON album_posts(album_id);
