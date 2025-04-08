ALTER TABLE tag_entities
ADD COLUMN tag_id UUID;

ALTER TABLE tag_entities
ADD CONSTRAINT fk_tag_entities_tag
FOREIGN KEY (tag_id) 
REFERENCES tags(id);
