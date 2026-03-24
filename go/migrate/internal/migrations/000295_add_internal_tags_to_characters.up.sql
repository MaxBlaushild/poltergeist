ALTER TABLE characters
ADD COLUMN internal_tags JSONB NOT NULL DEFAULT '[]'::jsonb;
