ALTER TABLE location_archetypes
  ADD COLUMN challenges_json jsonb NOT NULL DEFAULT '[]'::jsonb;

UPDATE location_archetypes
SET challenges_json = COALESCE(
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'question', challenge,
        'submissionType', COALESCE(submission_type, 'photo')
      )
    )
    FROM unnest(challenges) AS challenge
  ),
  '[]'::jsonb
);

ALTER TABLE location_archetypes
  DROP COLUMN challenges;

ALTER TABLE location_archetypes
  RENAME COLUMN challenges_json TO challenges;

ALTER TABLE location_archetypes
  DROP COLUMN submission_type;
