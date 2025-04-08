ALTER TABLE tag_entities
ADD CONSTRAINT fk_tag_entities_tag
FOREIGN KEY (tag_id) 
REFERENCES tags(id);
