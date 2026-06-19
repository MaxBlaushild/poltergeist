-- Optional photos attached to a mission submission. Stored inline (small,
-- phone-resized JPEGs) and served by the app — no external object store needed.
CREATE TABLE IF NOT EXISTS vampire_submission_photos (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  submission_id UUID NOT NULL REFERENCES vampire_mission_submissions(id) ON DELETE CASCADE,
  content_type TEXT NOT NULL DEFAULT 'image/jpeg',
  data BYTEA NOT NULL
);
CREATE INDEX IF NOT EXISTS vampire_submission_photos_submission_idx
  ON vampire_submission_photos(submission_id);
