CREATE TABLE user_challenge_completions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
  UNIQUE(user_id, challenge_id)
);

CREATE INDEX idx_user_challenge_completions_user_id ON user_challenge_completions(user_id);
CREATE INDEX idx_user_challenge_completions_challenge_id ON user_challenge_completions(challenge_id);
