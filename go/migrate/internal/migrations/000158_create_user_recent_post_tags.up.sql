CREATE TABLE user_recent_post_tags (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    last_posted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, tag)
);
CREATE INDEX idx_user_recent_post_tags_user_last ON user_recent_post_tags(user_id, last_posted_at DESC);
