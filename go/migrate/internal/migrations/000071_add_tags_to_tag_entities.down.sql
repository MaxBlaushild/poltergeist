ALTER TABLE tag_entities
DROP CONSTRAINT fk_tag_entities_tag;

ALTER TABLE tag_entities
DROP COLUMN tag_id;
