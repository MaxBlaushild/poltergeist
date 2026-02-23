ALTER TABLE location_archetypes
  ADD COLUMN challenges_text text[] NOT NULL DEFAULT '{}'::text[],
  ADD COLUMN submission_type text NOT NULL DEFAULT 'photo';

UPDATE location_archetypes
SET challenges_text = COALESCE(
  (
    SELECT ARRAY_AGG(elem->>'question')
    FROM jsonb_array_elements(challenges) AS elem
  ),
  '{}'::text[]
),
submission_type = COALESCE(
  (
    SELECT elem->>'submissionType'
    FROM jsonb_array_elements(challenges) AS elem
    LIMIT 1
  ),
  'photo'
);

ALTER TABLE location_archetypes
  DROP COLUMN challenges;

ALTER TABLE location_archetypes
  RENAME COLUMN challenges_text TO challenges;
