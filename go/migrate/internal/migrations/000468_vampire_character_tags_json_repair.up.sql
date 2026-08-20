-- One-time data repair for a job-runner bug: PetitionTheFount forces
-- JSON-object output on every completion (see fount-of-erebos's
-- jsonResponseInstruction/response_format=json_object), so the tag
-- generation prompt — which asked for a plain comma-separated list — came
-- back wrapped anyway, e.g. {"tags": "brilliant, unraveling, ..."}. The
-- naive comma-split in generate_character_tags.go then split that whole
-- blob into a tags array whose first element still carries the leading
-- `{"tags": "` and whose last element still carries the trailing `"}`,
-- e.g. ['{"tags": "brilliant', 'unraveling', ..., 'unsettling"}'] — which
-- is why it *looked* like one stringified-JSON tag when the admin UI
-- joined the array back into a comma-separated display string.
--
-- Strip those leftover fragments from any character whose tags array shows
-- this exact corruption (detected by its first element starting with
-- `{"tags":`). The processor itself is fixed separately (job-runner) to
-- request and parse real JSON going forward.
UPDATE vampire_characters c
SET tags = fixed.new_tags, updated_at = NOW()
FROM (
  SELECT
    c2.id,
    jsonb_agg(
      lower(trim(both ' ' from regexp_replace(regexp_replace(t.elem, '^\{"tags":\s*"', ''), '"\}\s*$', '')))
      ORDER BY t.i
    ) AS new_tags
  FROM vampire_characters c2
  CROSS JOIN LATERAL jsonb_array_elements_text(c2.tags) WITH ORDINALITY AS t(elem, i)
  WHERE jsonb_typeof(c2.tags) = 'array'
    AND jsonb_array_length(c2.tags) > 0
    AND (c2.tags ->> 0) LIKE '{"tags":%'
  GROUP BY c2.id
) AS fixed
WHERE c.id = fixed.id;
