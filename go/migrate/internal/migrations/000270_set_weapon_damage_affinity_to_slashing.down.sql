UPDATE inventory_items
SET damage_affinity = 'physical'
WHERE hand_item_category = 'weapon'
  AND damage_affinity = 'slashing';
