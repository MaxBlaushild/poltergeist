DELETE FROM reef_parameter_schemas
WHERE product_id = (SELECT id FROM reef_products WHERE slug = 'shelf-rack')
  AND generator_module = 'shelf_rack'
  AND version = 2;

UPDATE reef_parameter_schemas
SET active = true
WHERE product_id = (SELECT id FROM reef_products WHERE slug = 'shelf-rack')
  AND generator_module = 'shelf_rack'
  AND version = 1;
