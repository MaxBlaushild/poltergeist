CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_value ON tags(value);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_groups_name ON tag_groups(name);
