-- Preserve trip labels as ordered custom destinations before removing their
-- catalog. A later reapply must not leave provider UUIDs that can never resolve.
DO $$
BEGIN
  LOCK TABLE tc_stops, tc_destinations IN ACCESS EXCLUSIVE MODE;
  UPDATE tc_stops SET draft = jsonb_set(draft, '{destinations}', (
    SELECT coalesce(jsonb_agg((entry - 'id' - 'parentId' - 'metadata' - 'theme') ||
      '{"kind":"place"}'::jsonb ORDER BY position), '[]'::jsonb)
    FROM jsonb_array_elements(draft->'destinations') WITH ORDINALITY AS items(entry,position)
  )) WHERE jsonb_typeof(draft->'destinations') = 'array';
  UPDATE tc_stops SET published = jsonb_set(published, '{destinations}', (
    SELECT coalesce(jsonb_agg((entry - 'id' - 'parentId' - 'metadata' - 'theme') ||
      '{"kind":"place"}'::jsonb ORDER BY position), '[]'::jsonb)
    FROM jsonb_array_elements(published->'destinations') WITH ORDINALITY AS items(entry,position)
  )) WHERE jsonb_typeof(published->'destinations') = 'array';
  DROP TABLE tc_destination_usage;
  DROP TABLE tc_destination_terms;
  DROP TABLE tc_destinations;
END $$;
