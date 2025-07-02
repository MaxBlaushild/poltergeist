-- Remove the seeded monsters added in the up migration
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