CREATE TABLE IF NOT EXISTS vampire_instance_quiz_questions (
  instance_id UUID NOT NULL REFERENCES vampire_instances(id) ON DELETE CASCADE,
  question_id UUID NOT NULL REFERENCES vampire_quiz_questions(id) ON DELETE CASCADE,
  included BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (instance_id, question_id)
);
CREATE INDEX IF NOT EXISTS vampire_instance_quiz_questions_instance_idx ON vampire_instance_quiz_questions(instance_id);

ALTER TABLE vampire_instances DROP COLUMN IF EXISTS mystery_id;
ALTER TABLE vampire_quiz_questions DROP COLUMN IF EXISTS mystery_id;
ALTER TABLE vampire_secrets DROP COLUMN IF EXISTS beat_id;
ALTER TABLE vampire_secrets DROP COLUMN IF EXISTS mystery_id;
DROP TABLE IF EXISTS vampire_mystery_beats;
DROP TABLE IF EXISTS vampire_mysteries;
