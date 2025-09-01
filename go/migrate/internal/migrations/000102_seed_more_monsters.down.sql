-- Remove the seeded monsters and their actions added in the up migration
-- First remove monster actions (due to foreign key constraint)
DELETE FROM monster_actions WHERE monster_id IN (
    SELECT id FROM monsters WHERE name IN (
        'Kobold',
        'Giant Rat',
        'Dire Wolf',
        'Zombie',
        'Black Bear',
        'Hobgoblin',
        'Brown Bear',
        'Bugbear',
        'Owlbear',
        'Ogre',
        'Minotaur',
        'Basilisk',
        'Young Black Dragon',
        'Troll',
        'Drow'
    )
);

-- Then remove the monsters themselves
DELETE FROM monsters WHERE name IN (
    'Kobold',
    'Giant Rat',
    'Dire Wolf',
    'Zombie',
    'Black Bear',
    'Hobgoblin',
    'Brown Bear',
    'Bugbear',
    'Owlbear',
    'Ogre',
    'Minotaur',
    'Basilisk',
    'Young Black Dragon',
    'Troll',
    'Drow'
);