-- Seed additional monsters with diverse challenge ratings and types
-- First insert the basic monster data without actions/abilities
INSERT INTO monsters (
    name, size, type, subtype, alignment, armor_class, hit_points, hit_dice, speed,
    strength, dexterity, constitution, intelligence, wisdom, charisma,
    challenge_rating, experience_points, passive_perception,
    damage_resistances, damage_immunities, condition_immunities,
    languages, description, source
) VALUES 

-- CR 1/8 creatures
(
    'Kobold', 'Small', 'humanoid', 'reptilian', 'lawful evil', 12, 5, '2d6 - 2', 30,
    7, 15, 9, 8, 7, 8,
    0.125, 25, 8,
    NULL, NULL, NULL,
    ARRAY['Common', 'Draconic'],
    'Small reptilian humanoids that worship dragons and live in underground warrens.',
    'Monster Manual'
),
(
    'Giant Rat', 'Small', 'beast', NULL, 'unaligned', 12, 7, '2d6', 30,
    7, 15, 11, 2, 10, 4,
    0.125, 25, 10,
    NULL, NULL, NULL,
    NULL,
    'A filthy, aggressive rodent the size of a small dog.',
    'Monster Manual'
),

-- CR 1/4 creatures
(
    'Dire Wolf', 'Large', 'beast', NULL, 'unaligned', 14, 37, '5d10 + 15', 50,
    17, 15, 17, 3, 12, 7,
    1, 200, 13,
    NULL, NULL, NULL,
    NULL,
    'A larger, more ferocious version of its common cousin, often used as a mount by goblins and orcs.',
    'Monster Manual'
),
(
    'Zombie', 'Medium', 'undead', NULL, 'neutral evil', 8, 22, '3d8 + 9', 20,
    13, 6, 16, 3, 6, 5,
    0.25, 50, 8,
    NULL, ARRAY['poison'], ARRAY['poisoned'],
    NULL,
    'A shambling corpse animated by dark magic, driven by an insatiable hunger for living flesh.',
    'Monster Manual'
),

-- CR 1/2 creatures
(
    'Black Bear', 'Medium', 'beast', NULL, 'unaligned', 11, 19, '3d8 + 3', 40,
    15, 10, 13, 2, 12, 7,
    0.5, 100, 13,
    NULL, NULL, NULL,
    NULL,
    'A powerful omnivore with a keen sense of smell and surprising agility.',
    'Monster Manual'
),
(
    'Hobgoblin', 'Medium', 'humanoid', 'goblinoid', 'lawful evil', 18, 11, '2d8 + 2', 30,
    13, 12, 12, 10, 10, 9,
    0.5, 100, 10,
    NULL, NULL, NULL,
    ARRAY['Common', 'Goblin'],
    'Militaristic cousins of goblins, known for their discipline and tactical prowess.',
    'Monster Manual'
),

-- CR 1 creatures
(
    'Brown Bear', 'Large', 'beast', NULL, 'unaligned', 11, 34, '4d10 + 12', 40,
    19, 10, 16, 2, 13, 7,
    1, 200, 13,
    NULL, NULL, NULL,
    NULL,
    'A massive omnivore that can rear up on its hind legs to tower over most humanoids.',
    'Monster Manual'
),
(
    'Bugbear', 'Medium', 'humanoid', 'goblinoid', 'chaotic evil', 16, 27, '5d8 + 5', 30,
    15, 14, 13, 8, 11, 9,
    1, 200, 10,
    NULL, NULL, NULL,
    ARRAY['Common', 'Goblin'],
    'Hairy cousins of goblins and hobgoblins, known for their stealth and brute strength.',
    'Monster Manual'
),

-- CR 2-3 creatures
(
    'Owlbear', 'Large', 'monstrosity', NULL, 'unaligned', 13, 59, '7d10 + 21', 40,
    20, 12, 17, 3, 12, 7,
    3, 700, 13,
    NULL, NULL, NULL,
    NULL,
    'A cross between a giant owl and a bear, this creature is a deadly predator of the deep woods.',
    'Monster Manual'
),
(
    'Ogre', 'Large', 'giant', NULL, 'chaotic evil', 11, 59, '7d10 + 21', 40,
    19, 8, 16, 5, 7, 7,
    2, 450, 8,
    NULL, NULL, NULL,
    ARRAY['Common', 'Giant'],
    'Brutish giants that hunt other creatures for food, often working as mercenaries for more intelligent evil creatures.',
    'Monster Manual'
),
(
    'Minotaur', 'Large', 'monstrosity', NULL, 'chaotic evil', 14, 76, '9d10 + 27', 40,
    18, 11, 16, 6, 16, 9,
    3, 700, 17,
    NULL, NULL, NULL,
    ARRAY['Abyssal'],
    'A monstrous creature with the head of a bull and the body of a human, cursed to wander endless mazes.',
    'Monster Manual'
),
(
    'Basilisk', 'Medium', 'monstrosity', NULL, 'unaligned', 15, 52, '8d8 + 16', 20,
    16, 8, 15, 2, 8, 7,
    3, 700, 9,
    NULL, NULL, NULL,
    NULL,
    'A reptilian creature whose gaze can turn living beings to stone.',
    'Monster Manual'
),

-- CR 5+ creatures
(
    'Young Black Dragon', 'Large', 'dragon', 'chromatic', 'chaotic evil', 18, 127, '15d10 + 60', 40,
    19, 14, 19, 12, 11, 15,
    7, 2900, 16,
    NULL, ARRAY['acid'], NULL,
    ARRAY['Common', 'Draconic'],
    'A malevolent dragon that dwells in swamps and marshes, delighting in corruption and decay.',
    'Monster Manual'
),
(
    'Troll', 'Large', 'giant', NULL, 'chaotic evil', 15, 84, '8d10 + 40', 30,
    18, 13, 20, 7, 9, 7,
    5, 1800, 12,
    NULL, NULL, NULL,
    ARRAY['Giant'],
    'A fearsome giant with incredible regenerative abilities, capable of regrowing lost limbs.',
    'Monster Manual'
),
(
    'Drow', 'Medium', 'humanoid', 'elf', 'neutral evil', 15, 13, '3d8', 30,
    10, 14, 10, 11, 11, 12,
    0.25, 50, 12,
    NULL, NULL, NULL,
    ARRAY['Elvish', 'Undercommon'],
    'Dark elves that dwell in the Underdark, known for their cruelty and spider worship.',
    'Monster Manual'
);

-- Now insert the monster actions into the monster_actions table
-- Kobold actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 0, 'Dagger', 'Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 4 (1d4 + 2) piercing damage.', 4, '1d4+2', 'piercing', 5
FROM monsters WHERE name = 'Kobold';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach, range_long)
SELECT id, 'action', 1, 'Sling', 'Ranged Weapon Attack: +4 to hit, range 30/120 ft., one target. Hit: 4 (1d4 + 2) bludgeoning damage.', 4, '1d4+2', 'bludgeoning', 30, 120
FROM monsters WHERE name = 'Kobold';

-- Kobold special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Sunlight Sensitivity', 'While in sunlight, the kobold has disadvantage on attack rolls, as well as on Wisdom (Perception) checks that rely on sight.'
FROM monsters WHERE name = 'Kobold';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 1, 'Pack Tactics', 'The kobold has advantage on an attack roll against a creature if at least one of the kobold''s allies is within 5 feet of the creature and the ally isn''t incapacitated.'
FROM monsters WHERE name = 'Kobold';

-- Giant Rat actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 0, 'Bite', 'Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 4 (1d4 + 2) piercing damage.', 4, '1d4+2', 'piercing', 5
FROM monsters WHERE name = 'Giant Rat';

-- Giant Rat special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Keen Smell', 'The rat has advantage on Wisdom (Perception) checks that rely on smell.'
FROM monsters WHERE name = 'Giant Rat';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 1, 'Pack Tactics', 'The rat has advantage on an attack roll against a creature if at least one of the rat''s allies is within 5 feet of the creature and the ally isn''t incapacitated.'
FROM monsters WHERE name = 'Giant Rat';

-- Dire Wolf actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach, special_effects)
SELECT id, 'action', 0, 'Bite', 'Melee Weapon Attack: +5 to hit, reach 5 ft., one target. Hit: 10 (2d6 + 3) piercing damage. If the target is a creature, it must succeed on a DC 13 Strength saving throw or be knocked prone.', 5, '2d6+3', 'piercing', 5, 'DC 13 Strength save or prone'
FROM monsters WHERE name = 'Dire Wolf';

-- Dire Wolf special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Keen Hearing and Smell', 'The wolf has advantage on Wisdom (Perception) checks that rely on hearing or smell.'
FROM monsters WHERE name = 'Dire Wolf';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 1, 'Pack Tactics', 'The wolf has advantage on an attack roll against a creature if at least one of the wolf''s allies is within 5 feet of the creature and the ally isn''t incapacitated.'
FROM monsters WHERE name = 'Dire Wolf';

-- Zombie actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 0, 'Slam', 'Melee Weapon Attack: +3 to hit, reach 5 ft., one target. Hit: 4 (1d6 + 1) bludgeoning damage.', 3, '1d6+1', 'bludgeoning', 5
FROM monsters WHERE name = 'Zombie';

-- Zombie special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Undead Fortitude', 'If damage reduces the zombie to 0 hit points, it must make a Constitution saving throw with a DC of 5 + the damage taken, unless the damage is radiant or from a critical hit. On a success, the zombie drops to 1 hit point instead.'
FROM monsters WHERE name = 'Zombie';

-- Black Bear actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 0, 'Bite', 'Melee Weapon Attack: +3 to hit, reach 5 ft., one target. Hit: 5 (1d6 + 2) piercing damage.', 3, '1d6+2', 'piercing', 5
FROM monsters WHERE name = 'Black Bear';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 1, 'Claws', 'Melee Weapon Attack: +3 to hit, reach 5 ft., one target. Hit: 7 (2d4 + 2) slashing damage.', 3, '2d4+2', 'slashing', 5
FROM monsters WHERE name = 'Black Bear';

-- Black Bear special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Keen Smell', 'The bear has advantage on Wisdom (Perception) checks that rely on smell.'
FROM monsters WHERE name = 'Black Bear';

-- Hobgoblin actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 0, 'Longsword', 'Melee Weapon Attack: +3 to hit, reach 5 ft., one target. Hit: 5 (1d8 + 1) slashing damage, or 6 (1d10 + 1) slashing damage if used with two hands.', 3, '1d8+1', 'slashing', 5
FROM monsters WHERE name = 'Hobgoblin';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach, range_long)
SELECT id, 'action', 1, 'Longbow', 'Ranged Weapon Attack: +3 to hit, range 150/600 ft., one target. Hit: 5 (1d8 + 1) piercing damage.', 3, '1d8+1', 'piercing', 150, 600
FROM monsters WHERE name = 'Hobgoblin';

-- Hobgoblin special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Martial Advantage', 'Once per turn, the hobgoblin can deal an extra 7 (2d6) damage to a creature it hits with a weapon attack if that creature is within 5 feet of an ally that isn''t incapacitated.'
FROM monsters WHERE name = 'Hobgoblin';

-- Brown Bear actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 0, 'Bite', 'Melee Weapon Attack: +6 to hit, reach 5 ft., one target. Hit: 8 (1d8 + 4) piercing damage.', 6, '1d8+4', 'piercing', 5
FROM monsters WHERE name = 'Brown Bear';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 1, 'Claws', 'Melee Weapon Attack: +6 to hit, reach 5 ft., one target. Hit: 11 (2d6 + 4) slashing damage.', 6, '2d6+4', 'slashing', 5
FROM monsters WHERE name = 'Brown Bear';

-- Brown Bear special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Keen Smell', 'The bear has advantage on Wisdom (Perception) checks that rely on smell.'
FROM monsters WHERE name = 'Brown Bear';

-- Bugbear actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 0, 'Morningstar', 'Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 6 (1d8 + 2) piercing damage.', 4, '1d8+2', 'piercing', 5
FROM monsters WHERE name = 'Bugbear';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach, range_long)
SELECT id, 'action', 1, 'Javelin', 'Melee or Ranged Weapon Attack: +4 to hit, reach 5 ft. or range 30/120 ft., one target. Hit: 5 (1d6 + 2) piercing damage.', 4, '1d6+2', 'piercing', 30, 120
FROM monsters WHERE name = 'Bugbear';

-- Bugbear special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Brute', 'A melee weapon deals one extra die of its damage when the bugbear hits with it (included in the attack).'
FROM monsters WHERE name = 'Bugbear';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 1, 'Surprise Attack', 'If the bugbear surprises a creature and hits it with an attack during the first round of combat, the target takes an extra 7 (2d6) damage from the attack.'
FROM monsters WHERE name = 'Bugbear';

-- Owlbear actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'action', 0, 'Multiattack', 'The owlbear makes two attacks: one with its beak and one with its claws.'
FROM monsters WHERE name = 'Owlbear';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 1, 'Beak', 'Melee Weapon Attack: +7 to hit, reach 5 ft., one creature. Hit: 10 (1d10 + 5) piercing damage.', 7, '1d10+5', 'piercing', 5
FROM monsters WHERE name = 'Owlbear';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 2, 'Claws', 'Melee Weapon Attack: +7 to hit, reach 5 ft., one target. Hit: 14 (2d8 + 5) slashing damage.', 7, '2d8+5', 'slashing', 5
FROM monsters WHERE name = 'Owlbear';

-- Owlbear special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Keen Sight and Smell', 'The owlbear has advantage on Wisdom (Perception) checks that rely on sight or smell.'
FROM monsters WHERE name = 'Owlbear';

-- Ogre actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 0, 'Greatclub', 'Melee Weapon Attack: +6 to hit, reach 5 ft., one target. Hit: 13 (2d8 + 4) bludgeoning damage.', 6, '2d8+4', 'bludgeoning', 5
FROM monsters WHERE name = 'Ogre';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach, range_long)
SELECT id, 'action', 1, 'Javelin', 'Melee or Ranged Weapon Attack: +6 to hit, reach 5 ft. or range 30/120 ft., one target. Hit: 11 (2d6 + 4) piercing damage.', 6, '2d6+4', 'piercing', 30, 120
FROM monsters WHERE name = 'Ogre';

-- Minotaur actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 0, 'Greataxe', 'Melee Weapon Attack: +6 to hit, reach 5 ft., one target. Hit: 17 (2d12 + 4) slashing damage.', 6, '2d12+4', 'slashing', 5
FROM monsters WHERE name = 'Minotaur';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 1, 'Gore', 'Melee Weapon Attack: +6 to hit, reach 5 ft., one target. Hit: 13 (2d8 + 4) piercing damage.', 6, '2d8+4', 'piercing', 5
FROM monsters WHERE name = 'Minotaur';

-- Minotaur special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Charge', 'If the minotaur moves at least 10 feet straight toward a target and then hits it with a gore attack on the same turn, the target takes an extra 9 (2d8) piercing damage. If the target is a creature, it must succeed on a DC 14 Strength saving throw or be pushed up to 10 feet away and knocked prone.'
FROM monsters WHERE name = 'Minotaur';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 1, 'Labyrinthine Recall', 'The minotaur can perfectly recall any path it has traveled.'
FROM monsters WHERE name = 'Minotaur';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 2, 'Reckless', 'At the start of its turn, the minotaur can gain advantage on all melee weapon attack rolls during that turn, but attack rolls against it have advantage until the start of its next turn.'
FROM monsters WHERE name = 'Minotaur';

-- Basilisk actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, additional_damage_dice, additional_damage_type, range_reach)
SELECT id, 'action', 0, 'Bite', 'Melee Weapon Attack: +5 to hit, reach 5 ft., one target. Hit: 10 (2d6 + 3) piercing damage plus 7 (2d6) poison damage.', 5, '2d6+3', 'piercing', '2d6', 'poison', 5
FROM monsters WHERE name = 'Basilisk';

-- Basilisk special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Petrifying Gaze', 'If a creature starts its turn within 30 feet of the basilisk and the two of them can see each other, the basilisk can force the creature to make a DC 12 Constitution saving throw if the basilisk isn''t incapacitated. On a failed save, the creature magically begins to turn to stone and is restrained. It must repeat the saving throw at the end of its next turn. On a success, the effect ends. On a failure, the creature is petrified until freed by the greater restoration spell or other magic.'
FROM monsters WHERE name = 'Basilisk';

-- Young Black Dragon actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'action', 0, 'Multiattack', 'The dragon makes three attacks: one with its bite and two with its claws.'
FROM monsters WHERE name = 'Young Black Dragon';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, additional_damage_dice, additional_damage_type, range_reach)
SELECT id, 'action', 1, 'Bite', 'Melee Weapon Attack: +7 to hit, reach 10 ft., one target. Hit: 15 (2d10 + 4) piercing damage plus 4 (1d8) acid damage.', 7, '2d10+4', 'piercing', '1d8', 'acid', 10
FROM monsters WHERE name = 'Young Black Dragon';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 2, 'Claw', 'Melee Weapon Attack: +7 to hit, reach 5 ft., one target. Hit: 11 (2d6 + 4) slashing damage.', 7, '2d6+4', 'slashing', 5
FROM monsters WHERE name = 'Young Black Dragon';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, save_dc, save_ability, damage_dice, damage_type, area_type, area_size, recharge)
SELECT id, 'action', 3, 'Acid Breath', 'The dragon exhales acid in a 30-foot line that is 5 feet wide. Each creature in that area must make a DC 14 Dexterity saving throw, taking 49 (11d8) acid damage on a failed save, or half as much damage on a successful one.', 14, 'Dexterity', '11d8', 'acid', 'line', 30, '5-6'
FROM monsters WHERE name = 'Young Black Dragon';

-- Young Black Dragon special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Amphibious', 'The dragon can breathe air and water.'
FROM monsters WHERE name = 'Young Black Dragon';

-- Troll actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'action', 0, 'Multiattack', 'The troll makes three attacks: one with its bite and two with its claws.'
FROM monsters WHERE name = 'Troll';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 1, 'Bite', 'Melee Weapon Attack: +7 to hit, reach 5 ft., one target. Hit: 7 (1d6 + 4) piercing damage.', 7, '1d6+4', 'piercing', 5
FROM monsters WHERE name = 'Troll';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 2, 'Claw', 'Melee Weapon Attack: +7 to hit, reach 5 ft., one target. Hit: 11 (2d6 + 4) slashing damage.', 7, '2d6+4', 'slashing', 5
FROM monsters WHERE name = 'Troll';

-- Troll special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Keen Smell', 'The troll has advantage on Wisdom (Perception) checks that rely on smell.'
FROM monsters WHERE name = 'Troll';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 1, 'Regeneration', 'The troll regains 10 hit points at the start of its turn. If the troll takes acid or fire damage, this trait doesn''t function at the start of the troll''s next turn. The troll dies only if it starts its turn with 0 hit points and doesn''t regenerate.'
FROM monsters WHERE name = 'Troll';

-- Drow actions
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach)
SELECT id, 'action', 0, 'Shortsword', 'Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6 + 2) piercing damage.', 4, '1d6+2', 'piercing', 5
FROM monsters WHERE name = 'Drow';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description, attack_bonus, damage_dice, damage_type, range_reach, range_long, special_effects)
SELECT id, 'action', 1, 'Hand Crossbow', 'Ranged Weapon Attack: +4 to hit, range 30/120 ft., one target. Hit: 5 (1d6 + 2) piercing damage, and the target must succeed on a DC 13 Constitution saving throw or be poisoned for 1 hour. If the saving throw fails by 5 or more, the target is also unconscious while poisoned in this way.', 4, '1d6+2', 'piercing', 30, 120, 'DC 13 Constitution save or poisoned'
FROM monsters WHERE name = 'Drow';

-- Drow special abilities
INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 0, 'Fey Ancestry', 'The drow has advantage on saving throws against being charmed, and magic can''t put the drow to sleep.'
FROM monsters WHERE name = 'Drow';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 1, 'Innate Spellcasting', 'The drow''s spellcasting ability is Charisma (spell save DC 11). It can innately cast the following spells, requiring no material components: At will: dancing lights. 1/day each: darkness, faerie fire.'
FROM monsters WHERE name = 'Drow';

INSERT INTO monster_actions (monster_id, action_type, order_index, name, description)
SELECT id, 'special_ability', 2, 'Sunlight Sensitivity', 'While in sunlight, the drow has disadvantage on attack rolls, as well as on Wisdom (Perception) checks that rely on sight.'
FROM monsters WHERE name = 'Drow';