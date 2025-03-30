INSERT INTO tag_groups (id, name, created_at, updated_at) VALUES 
    (gen_random_uuid(), 'sightseeing', NOW(), NOW()),
    (gen_random_uuid(), 'shopping', NOW(), NOW()),
    (gen_random_uuid(), 'eating', NOW(), NOW()),
    (gen_random_uuid(), 'drinking', NOW(), NOW()),
    (gen_random_uuid(), 'fitness', NOW(), NOW());

WITH shopping_group AS (
    SELECT id FROM tag_groups WHERE name = 'shopping'
)
INSERT INTO tags (id, tag_group_id, value, created_at, updated_at) VALUES
    (gen_random_uuid(), (SELECT id FROM shopping_group), 'clothes', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM shopping_group), 'games', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM shopping_group), 'toys', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM shopping_group), 'jewlery', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM shopping_group), 'gifts', NOW(), NOW());

WITH sightseeing_group AS (
    SELECT id FROM tag_groups WHERE name = 'sightseeing'
)
INSERT INTO tags (id, tag_group_id, value, created_at, updated_at) VALUES
    (gen_random_uuid(), (SELECT id FROM sightseeing_group), 'museums', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM sightseeing_group), 'historical', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM sightseeing_group), 'nature', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM sightseeing_group), 'architecture', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM sightseeing_group), 'art', NOW(), NOW());

WITH eating_group AS (
    SELECT id FROM tag_groups WHERE name = 'eating'
)
INSERT INTO tags (id, tag_group_id, value, created_at, updated_at) VALUES
    (gen_random_uuid(), (SELECT id FROM eating_group), 'italian', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'japanese', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'korean', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'chinese', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'american', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'mexican', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'indian', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'thai', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'vietnamese', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'brazilian', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'peruvian', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'moroccan', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM eating_group), 'turkish', NOW(), NOW());

WITH drinking_group AS (
    SELECT id FROM tag_groups WHERE name = 'drinking'
)
INSERT INTO tags (id, tag_group_id, value, created_at, updated_at) VALUES
    (gen_random_uuid(), (SELECT id FROM drinking_group), 'coffee', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM drinking_group), 'tea', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM drinking_group), 'cocktails', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM drinking_group), 'wines', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM drinking_group), 'beers', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM drinking_group), 'spirits', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM drinking_group), 'soft_drinks', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM drinking_group), 'alcoholic_drinks', NOW(), NOW());

WITH fitness_group AS (
    SELECT id FROM tag_groups WHERE name = 'fitness'
)
INSERT INTO tags (id, tag_group_id, value, created_at, updated_at) VALUES
    (gen_random_uuid(), (SELECT id FROM fitness_group), 'gym', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM fitness_group), 'yoga', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM fitness_group), 'pilates', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM fitness_group), 'crossfit', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM fitness_group), 'strength_training', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM fitness_group), 'cardio', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM fitness_group), 'hiking', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM fitness_group), 'cycling', NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM fitness_group), 'swimming', NOW(), NOW());

        