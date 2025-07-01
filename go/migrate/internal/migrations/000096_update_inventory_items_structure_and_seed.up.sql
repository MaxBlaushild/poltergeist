-- Update inventory_items table structure to match the hardcoded items
ALTER TABLE inventory_items DROP CONSTRAINT inventory_items_pkey;
ALTER TABLE inventory_items DROP COLUMN id;
ALTER TABLE inventory_items ADD COLUMN id SERIAL PRIMARY KEY;
ALTER TABLE inventory_items ADD COLUMN rarity_tier TEXT NOT NULL DEFAULT 'Common';
ALTER TABLE inventory_items ADD COLUMN is_capture_type BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE inventory_items ADD COLUMN item_type TEXT NOT NULL DEFAULT 'consumable';
ALTER TABLE inventory_items ADD COLUMN equipment_slot TEXT;

-- Update owned_inventory_items to use INTEGER instead of UUID for inventory_item_id
-- (This should already be correct based on the existing model)

-- Seed the table with the current hardcoded items
INSERT INTO inventory_items (id, created_at, updated_at, name, image_url, flavor_text, effect_text, rarity_tier, is_capture_type, item_type, equipment_slot) VALUES
(1, NOW(), NOW(), 'Cipher of the Laughing Monkey', 'https://crew-points-of-interest.s3.amazonaws.com/cipher.png', 'Unearthed in the heart of a dense jungle, this mysterious item lay among countless laughing skeletons.', 'Deploy to sow confusion among your rivals by warping their clue texts into bewildering riddles.', 'Uncommon', false, 'consumable', NULL),
(2, NOW(), NOW(), 'Golden Telescope', 'https://crew-points-of-interest.s3.amazonaws.com/telescope-better.png', 'Legend has it that a artificer parted with his sight to create this so that others might see the stars.', 'Instantly reveals a hidden point on the map. Tap this icon next to the "I''m here!" button on a hidden points of interest to use it.', 'Uncommon', false, 'consumable', NULL),
(3, NOW(), NOW(), 'Flawed Ruby', 'https://crew-points-of-interest.s3.amazonaws.com/flawed-ruby.png', 'This gem is chipped and disfigured, but will still fetch a decent price at market.', 'Instantly captures a tier one challenge. Tap this icon next to the "Submit Answer" button on any unlocked tier one challenge to use it.', 'Uncommon', true, 'consumable', NULL),
(4, NOW(), NOW(), 'Ruby', 'https://crew-points-of-interest.s3.amazonaws.com/ruby.png', 'A gem, sparkling more red than the blood you had to spill to procure it.', 'Instantly captures a tier two challenge. Tap this icon next to the "Submit Answer" button on any unlocked tier two challenge to use it.', 'Epic', true, 'consumable', NULL),
(5, NOW(), NOW(), 'Brilliant Ruby', 'https://crew-points-of-interest.s3.amazonaws.com/brilliant-ruby.png', 'You''ve hit the motherload! This gem will fetch a pirate''s ransom.', 'Instantly captures a tier three challenge. Tap this icon next to the "Submit Answer" button on any unlocked tier three challenge to use it.', 'Mythic', true, 'consumable', NULL),
(6, NOW(), NOW(), 'Cortez''s Cutlass', 'https://crew-points-of-interest.s3.amazonaws.com/cortez-cutlass.png', 'A relic of the high seas, its blade still sharp enough to cut through the thickest of hides.', 'Steal all of another team''s items.', 'Not Droppable', false, 'equippable', 'right_hand'),
(7, NOW(), NOW(), 'Rusted Musket', 'https://crew-points-of-interest.s3.amazonaws.com/rusted-musket.png', 'Found in a shipwreck, its barrel rusted and its stock worn.', 'Use on an opponent to lower their score by 2.', 'Common', false, 'consumable', NULL),
(8, NOW(), NOW(), 'Gold Coin', 'https://crew-points-of-interest.s3.amazonaws.com/gold-coin.png', 'A coin of pure gold. The currency of the high seas.', 'Hold in your inventory to increase your score by 1.', 'Common', false, 'passive', NULL),
(9, NOW(), NOW(), 'Dagger', 'https://crew-points-of-interest.s3.amazonaws.com/dagger.png', 'A small, sharp blade. It''s not much, but it''s better than nothing.', 'Steal one item from an opponent at random.', 'Epic', false, 'equippable', 'left_hand'),
(10, NOW(), NOW(), 'Damage', 'https://crew-points-of-interest.s3.amazonaws.com/bullet-hole.png', 'You''ve been shot! Some ale will help.', 'Decreases score by 2 while held in inventory.', 'Not Droppable', false, 'passive', NULL),
(11, NOW(), NOW(), 'Entseed', 'https://crew-points-of-interest.s3.amazonaws.com/entseed.png', 'This seed will grow into an Ent one day. For now, you can just bask in it''s life energy.', 'Increase score by 3 and neutralize the effects of Damage while held in inventory.', 'Not Droppable', false, 'passive', NULL),
(12, NOW(), NOW(), 'Ale', 'https://crew-points-of-interest.s3.amazonaws.com/ale.png', 'A hearty brew, made from the finest ingredients.', 'Removes one damage when drank.', 'Uncommon', false, 'consumable', NULL),
(13, NOW(), NOW(), 'Witchflame', 'https://crew-points-of-interest.s3.us-east-1.amazonaws.com/witchflame.png', 'A flame that burns with a sinister glow.', 'Removes all damage when held. Also increases score by 1 when held.', 'Not Droppable', false, 'passive', NULL),
(14, NOW(), NOW(), 'Wicked Spellbook', 'https://crew-points-of-interest.s3.us-east-1.amazonaws.com/wicked-spellbook.png', 'The spellbook whispers to you. Ignore it.', 'Steal all of another team''s items.', 'Not Droppable', false, 'equippable', 'left_hand'),
(15, NOW(), NOW(), 'The Compass of Peace', 'https://crew-points-of-interest.s3.us-east-1.amazonaws.com/compass-of-peace.png', 'Given to you by Shalimar the Merchant. The compass is said to point towards what the wearer needs most to heal.', 'Negate up to 3 damage when held.', 'Not Droppable', false, 'equippable', 'neck'),
(16, NOW(), NOW(), 'Pirate''s Tricorn Hat', 'https://crew-points-of-interest.s3.amazonaws.com/tricorn-hat.png', 'A weathered hat that has seen many adventures on the high seas. Its feathers still dance in the wind.', 'Increases treasure finding by 10% when worn.', 'Uncommon', false, 'equippable', 'head'),
(17, NOW(), NOW(), 'Captain''s Coat', 'https://crew-points-of-interest.s3.amazonaws.com/captains-coat.png', 'A noble coat worn by a captain of renown. Its brass buttons still gleam despite the salt and spray.', 'Provides +5 defense against damage when worn.', 'Epic', false, 'equippable', 'chest'),
(18, NOW(), NOW(), 'Seafarer''s Boots', 'https://crew-points-of-interest.s3.amazonaws.com/seafarer-boots.png', 'Sturdy boots made for walking on both deck and shore. They''ve never failed their wearer.', 'Increases movement speed by 15% when worn.', 'Common', false, 'equippable', 'feet'),
(19, NOW(), NOW(), 'Enchanted Ring of Fortune', 'https://crew-points-of-interest.s3.amazonaws.com/fortune-ring.png', 'A mysterious ring that pulses with magical energy. Those who wear it speak of incredible luck.', 'Doubles reward chances for treasure hunting when worn.', 'Mythic', false, 'equippable', 'ring'),
(20, NOW(), NOW(), 'Leather Sailing Gloves', 'https://crew-points-of-interest.s3.amazonaws.com/sailing-gloves.png', 'Well-worn gloves that have handled countless ropes and rigging. They fit like a second skin.', 'Improves grip and handling, reducing chance of dropping items by 50%.', 'Common', false, 'equippable', 'gloves');

-- Reset the sequence for the id column to start from 21 for new items
SELECT setval('inventory_items_id_seq', (SELECT MAX(id) FROM inventory_items) + 1);