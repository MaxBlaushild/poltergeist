ALTER TABLE quest_nodes
  ADD COLUMN submission_type TEXT;

UPDATE quest_nodes
SET submission_type = 'photo'
WHERE submission_type IS NULL OR submission_type = '';

ALTER TABLE quest_nodes
  ALTER COLUMN submission_type SET NOT NULL,
  ALTER COLUMN submission_type SET DEFAULT 'photo';
