CREATE TABLE feedback_items (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  zone_id UUID REFERENCES zones(id) ON DELETE SET NULL,
  route TEXT NOT NULL DEFAULT '',
  message TEXT NOT NULL
);

CREATE INDEX idx_feedback_items_created_at
  ON feedback_items(created_at DESC);

CREATE INDEX idx_feedback_items_user_id
  ON feedback_items(user_id);

CREATE INDEX idx_feedback_items_zone_id
  ON feedback_items(zone_id);
