INSERT INTO inventory_items (id, name, image_url, flavor_text, effect_text, rarity_tier, is_capture_type, created_at, updated_at) VALUES
(1, 'Cipher of the Laughing Monkey', 'https://crew-points-of-interest.s3.amazonaws.com/cipher.png', 'Unearthed in the heart of a dense jungle, this mysterious item lay among countless laughing skeletons.', 'Deploy to sow confusion among your rivals by warping their clue texts into bewildering riddles.', 'Uncommon', FALSE, NOW(), NOW()),
(2, 'Golden Telescope', 'https://crew-points-of-interest.s3.amazonaws.com/telescope-better.png', 'Legend has it that a artificer parted with his sight to create this so that others might see the stars.', 'Instantly reveals a hidden point on the map. Tap this icon next to the "I''m here!" button on a hidden points of interest to use it.', 'Uncommon', FALSE, NOW(), NOW()),
(3, 'Flawed Ruby', 'https://crew-points-of-interest.s3.amazonaws.com/flawed-ruby.png', 'This gem is chipped and disfigured, but will still fetch a decent price at market.', 'Instantly captures a tier one challenge. Tap this icon next to the "Submit Answer" button on any unlocked tier one challenge to use it.', 'Uncommon', TRUE, NOW(), NOW()),
(4, 'Ruby', 'https://crew-points-of-interest.s3.amazonaws.com/ruby.png', 'A gem, sparkling more red than the blood you had to spill to procure it.', 'Instantly captures a tier two challenge. Tap this icon next to the "Submit Answer" button on any unlocked tier two challenge to use it.', 'Epic', TRUE, NOW(), NOW()),
(5, 'Brilliant Ruby', 'https://crew-points-of-interest.s3.amazonaws.com/brilliant-ruby.png', 'You''ve hit the motherload! This gem will fetch a pirate''s ransom.', 'Instantly captures a tier three challenge. Tap this icon next to the "Submit Answer" button on any unlocked tier three challenge to use it.', 'Mythic', TRUE, NOW(), NOW()),
(6, 'Cortez''s Cutlass', 'https://crew-points-of-interest.s3.amazonaws.com/cortez-cutlass.png', 'A relic of the high seas, its blade still sharp enough to cut through the thickest of hides.', 'Steal all of another team''s items.', 'Not Droppable', FALSE, NOW(), NOW()),
(7, 'Rusted Musket', 'https://crew-points-of-interest.s3.amazonaws.com/rusted-musket.png', 'Found in a shipwreck, its barrel rusted and its stock worn.', 'Use on an opponent to lower their score by 2.', 'Common', FALSE, NOW(), NOW()),
(8, 'Gold Coin', 'https://crew-points-of-interest.s3.amazonaws.com/gold-coin.png', 'A coin of pure gold. The currency of the high seas.', 'Hold in your inventory to increase your score by 1.', 'Common', FALSE, NOW(), NOW()),
(9, 'Dagger', 'https://crew-points-of-interest.s3.amazonaws.com/dagger.png', 'A small, sharp blade. It''s not much, but it''s better than nothing.', 'Steal one item from an opponent at random.', 'Epic', FALSE, NOW(), NOW()),
(10, 'Damage', 'https://crew-points-of-interest.s3.amazonaws.com/bullet-hole.png', 'You''ve been shot! Some ale will help.', 'Decreases score by 2 while held in inventory.', 'Not Droppable', FALSE, NOW(), NOW()),
(11, 'Entseed', 'https://crew-points-of-interest.s3.amazonaws.com/entseed.png', 'This seed will grow into an Ent one day. For now, you can just bask in it''s life energy.', 'Increase score by 3 and neutralize the effects of Damage while held in inventory.', 'Not Droppable', FALSE, NOW(), NOW()),
(12, 'Ale', 'https://crew-points-of-interest.s3.amazonaws.com/ale.png', 'A hearty brew, made from the finest ingredients.', 'Removes one damage when drank.', 'Uncommon', FALSE, NOW(), NOW()),
(13, 'Witchflame', 'https://crew-points-of-interest.s3.us-east-1.amazonaws.com/witchflame.png', 'A flame that burns with a sinister glow.', 'Removes all damage when held. Also increases score by 1 when held.', 'Not Droppable', FALSE, NOW(), NOW()),
(14, 'Wicked Spellbook', 'https://crew-points-of-interest.s3.us-east-1.amazonaws.com/wicked-spellbook.png', 'The spellbook whispers to you. Ignore it.', 'Steal all of another team''s items.', 'Not Droppable', FALSE, NOW(), NOW()),
(15, 'The Compass of Peace', 'https://crew-points-of-interest.s3.us-east-1.amazonaws.com/compass-of-peace.png', 'Given to you by Shalimar the Merchant. The compass is said to point towards what the wearer needs most to heal.', 'Negate up to 3 damage when held.', 'Not Droppable', FALSE, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Reset the sequence to the max ID to ensure next inserts get proper IDs
-- Only set the sequence if it exists (it should be created automatically by SERIAL)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'inventory_items_id_seq') THEN
        PERFORM setval('inventory_items_id_seq', COALESCE((SELECT MAX(id) FROM inventory_items), 1), false);
    END IF;
END $$;

