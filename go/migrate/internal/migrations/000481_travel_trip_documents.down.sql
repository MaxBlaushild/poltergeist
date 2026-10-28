-- Remove uploaded files through the API (which also removes their S3 objects)
-- before rollback. Dropping the only object references would orphan originals.
DO $$ BEGIN
 IF EXISTS (SELECT 1 FROM tc_trip_documents WHERE storage_key <> '') THEN
   RAISE EXCEPTION 'Remove uploaded trip documents through the API before rolling back migration 481';
 END IF;
END $$;
DROP TABLE IF EXISTS tc_trip_document_comments;
DROP TABLE IF EXISTS tc_trip_documents;
