DO $$
BEGIN
  LOCK TABLE documents IN ACCESS EXCLUSIVE MODE;
  IF EXISTS (SELECT 1 FROM documents WHERE visibility <> 'network') THEN
    RAISE EXCEPTION 'Travel Library downgrade blocked: restricted documents remain; explicitly migrate private and selected documents before downgrading'
      USING ERRCODE = 'check_violation',
            HINT = 'Explicitly migrate private and selected documents before downgrading; their access restrictions cannot be preserved by the older schema.';
  END IF;
END;
$$;

DROP TABLE document_trip_references;
DROP TABLE document_saves;
DROP TABLE document_shares;
DROP INDEX documents_library_audience_idx;
ALTER TABLE documents DROP COLUMN destination, DROP COLUMN personal_note, DROP COLUMN format, DROP COLUMN visibility;
DELETE FROM trending_destinations;
