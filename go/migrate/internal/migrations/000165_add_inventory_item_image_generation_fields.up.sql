ALTER TABLE inventory_items
ADD COLUMN image_generation_status TEXT NOT NULL DEFAULT 'none',
ADD COLUMN image_generation_error TEXT;

UPDATE inventory_items
SET image_generation_status = 'complete'
WHERE image_url IS NOT NULL AND image_url <> '';
