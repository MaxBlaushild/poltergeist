-- Seed additional monsters with diverse challenge ratings and types
INSERT INTO monsters (
    name, size, type, subtype, alignment, armor_class, hit_points, hit_dice, speed,
    strength, dexterity, constitution, intelligence, wisdom, charisma,
    challenge_rating, experience_points, passive_perception,
    damage_resistances, damage_immunities, condition_immunities,
    languages, actions, special_abilities, description, source
) VALUES 

-- CR 1/8 creatures
(
    'Kobold', 'Small', 'humanoid', 'reptilian', 'lawful evil', 12, 5, '2d6 - 2', 30,
    7, 15, 9, 8, 7, 8,
    0.125, 25, 8,
    NULL, NULL, NULL,
    ARRAY['Common', 'Draconic'],
    '[{"name": "Dagger", "description": "Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 4 (1d4 + 2) piercing damage.", "attack_bonus": 4, "damage": "1d4+2", "damage_type": "piercing"}, {"name": "Sling", "description": "Ranged Weapon Attack: +4 to hit, range 30/120 ft., one target. Hit: 4 (1d4 + 2) bludgeoning damage.", "attack_bonus": 4, "damage": "1d4+2", "damage_type": "bludgeoning"}]',
    '[{"name": "Sunlight Sensitivity", "description": "While in sunlight, the kobold has disadvantage on attack rolls, as well as on Wisdom (Perception) checks that rely on sight."}, {"name": "Pack Tactics", "description": "The kobold has advantage on an attack roll against a creature if at least one of the kobold''s allies is within 5 feet of the creature and the ally isn''t incapacitated."}]',
    'Small reptilian humanoids that worship dragons and live in underground warrens.',
    'Monster Manual'
),
(
    'Giant Rat', 'Small', 'beast', NULL, 'unaligned', 12, 7, '2d6', 30,
    7, 15, 11, 2, 10, 4,
    0.125, 25, 10,
    NULL, NULL, NULL,
    NULL,
    '[{"name": "Bite", "description": "Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 4 (1d4 + 2) piercing damage.", "attack_bonus": 4, "damage": "1d4+2", "damage_type": "piercing"}]',
    '[{"name": "Keen Smell", "description": "The rat has advantage on Wisdom (Perception) checks that rely on smell."}, {"name": "Pack Tactics", "description": "The rat has advantage on an attack roll against a creature if at least one of the rat''s allies is within 5 feet of the creature and the ally isn''t incapacitated."}]',
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
    '[{"name": "Bite", "description": "Melee Weapon Attack: +5 to hit, reach 5 ft., one target. Hit: 10 (2d6 + 3) piercing damage. If the target is a creature, it must succeed on a DC 13 Strength saving throw or be knocked prone.", "attack_bonus": 5, "damage": "2d6+3", "damage_type": "piercing", "special": "DC 13 Strength save or prone"}]',
    '[{"name": "Keen Hearing and Smell", "description": "The wolf has advantage on Wisdom (Perception) checks that rely on hearing or smell."}, {"name": "Pack Tactics", "description": "The wolf has advantage on an attack roll against a creature if at least one of the wolf''s allies is within 5 feet of the creature and the ally isn''t incapacitated."}]',
    'A larger, more ferocious version of its common cousin, often used as a mount by goblins and orcs.',
    'Monster Manual'
),
(
    'Zombie', 'Medium', 'undead', NULL, 'neutral evil', 8, 22, '3d8 + 9', 20,
    13, 6, 16, 3, 6, 5,
    0.25, 50, 8,
    NULL, ARRAY['poison'], ARRAY['poisoned'],
    NULL,
    '[{"name": "Slam", "description": "Melee Weapon Attack: +3 to hit, reach 5 ft., one target. Hit: 4 (1d6 + 1) bludgeoning damage.", "attack_bonus": 3, "damage": "1d6+1", "damage_type": "bludgeoning"}]',
    '[{"name": "Undead Fortitude", "description": "If damage reduces the zombie to 0 hit points, it must make a Constitution saving throw with a DC of 5 + the damage taken, unless the damage is radiant or from a critical hit. On a success, the zombie drops to 1 hit point instead."}]',
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
    '[{"name": "Bite", "description": "Melee Weapon Attack: +3 to hit, reach 5 ft., one target. Hit: 5 (1d6 + 2) piercing damage.", "attack_bonus": 3, "damage": "1d6+2", "damage_type": "piercing"}, {"name": "Claws", "description": "Melee Weapon Attack: +3 to hit, reach 5 ft., one target. Hit: 7 (2d4 + 2) slashing damage.", "attack_bonus": 3, "damage": "2d4+2", "damage_type": "slashing"}]',
    '[{"name": "Keen Smell", "description": "The bear has advantage on Wisdom (Perception) checks that rely on smell."}]',
    'A powerful omnivore with a keen sense of smell and surprising agility.',
    'Monster Manual'
),
(
    'Hobgoblin', 'Medium', 'humanoid', 'goblinoid', 'lawful evil', 18, 11, '2d8 + 2', 30,
    13, 12, 12, 10, 10, 9,
    0.5, 100, 10,
    NULL, NULL, NULL,
    ARRAY['Common', 'Goblin'],
    '[{"name": "Longsword", "description": "Melee Weapon Attack: +3 to hit, reach 5 ft., one target. Hit: 5 (1d8 + 1) slashing damage, or 6 (1d10 + 1) slashing damage if used with two hands.", "attack_bonus": 3, "damage": "1d8+1", "damage_type": "slashing"}, {"name": "Longbow", "description": "Ranged Weapon Attack: +3 to hit, range 150/600 ft., one target. Hit: 5 (1d8 + 1) piercing damage.", "attack_bonus": 3, "damage": "1d8+1", "damage_type": "piercing"}]',
    '[{"name": "Martial Advantage", "description": "Once per turn, the hobgoblin can deal an extra 7 (2d6) damage to a creature it hits with a weapon attack if that creature is within 5 feet of an ally that isn''t incapacitated."}]',
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
    '[{"name": "Bite", "description": "Melee Weapon Attack: +6 to hit, reach 5 ft., one target. Hit: 8 (1d8 + 4) piercing damage.", "attack_bonus": 6, "damage": "1d8+4", "damage_type": "piercing"}, {"name": "Claws", "description": "Melee Weapon Attack: +6 to hit, reach 5 ft., one target. Hit: 11 (2d6 + 4) slashing damage.", "attack_bonus": 6, "damage": "2d6+4", "damage_type": "slashing"}]',
    '[{"name": "Keen Smell", "description": "The bear has advantage on Wisdom (Perception) checks that rely on smell."}]',
    'A massive omnivore that can rear up on its hind legs to tower over most humanoids.',
    'Monster Manual'
),
(
    'Bugbear', 'Medium', 'humanoid', 'goblinoid', 'chaotic evil', 16, 27, '5d8 + 5', 30,
    15, 14, 13, 8, 11, 9,
    1, 200, 10,
    NULL, NULL, NULL,
    ARRAY['Common', 'Goblin'],
    '[{"name": "Morningstar", "description": "Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 6 (1d8 + 2) piercing damage.", "attack_bonus": 4, "damage": "1d8+2", "damage_type": "piercing"}, {"name": "Javelin", "description": "Melee or Ranged Weapon Attack: +4 to hit, reach 5 ft. or range 30/120 ft., one target. Hit: 5 (1d6 + 2) piercing damage.", "attack_bonus": 4, "damage": "1d6+2", "damage_type": "piercing"}]',
    '[{"name": "Brute", "description": "A melee weapon deals one extra die of its damage when the bugbear hits with it (included in the attack)."}, {"name": "Surprise Attack", "description": "If the bugbear surprises a creature and hits it with an attack during the first round of combat, the target takes an extra 7 (2d6) damage from the attack."}]',
    'Hairy cousins of goblins and hobgoblins, known for their stealth and brute strength.',
    'Monster Manual'
),

-- CR 2 creatures
(
    'Owlbear', 'Large', 'monstrosity', NULL, 'unaligned', 13, 59, '7d10 + 21', 40,
    20, 12, 17, 3, 12, 7,
    3, 700, 13,
    NULL, NULL, NULL,
    NULL,
    '[{"name": "Multiattack", "description": "The owlbear makes two attacks: one with its beak and one with its claws."}, {"name": "Beak", "description": "Melee Weapon Attack: +7 to hit, reach 5 ft., one creature. Hit: 10 (1d10 + 5) piercing damage.", "attack_bonus": 7, "damage": "1d10+5", "damage_type": "piercing"}, {"name": "Claws", "description": "Melee Weapon Attack: +7 to hit, reach 5 ft., one target. Hit: 14 (2d8 + 5) slashing damage.", "attack_bonus": 7, "damage": "2d8+5", "damage_type": "slashing"}]',
    '[{"name": "Keen Sight and Smell", "description": "The owlbear has advantage on Wisdom (Perception) checks that rely on sight or smell."}]',
    'A cross between a giant owl and a bear, this creature is a deadly predator of the deep woods.',
    'Monster Manual'
),
(
    'Ogre', 'Large', 'giant', NULL, 'chaotic evil', 11, 59, '7d10 + 21', 40,
    19, 8, 16, 5, 7, 7,
    2, 450, 8,
    NULL, NULL, NULL,
    ARRAY['Common', 'Giant'],
    '[{"name": "Greatclub", "description": "Melee Weapon Attack: +6 to hit, reach 5 ft., one target. Hit: 13 (2d8 + 4) bludgeoning damage.", "attack_bonus": 6, "damage": "2d8+4", "damage_type": "bludgeoning"}, {"name": "Javelin", "description": "Melee or Ranged Weapon Attack: +6 to hit, reach 5 ft. or range 30/120 ft., one target. Hit: 11 (2d6 + 4) piercing damage.", "attack_bonus": 6, "damage": "2d6+4", "damage_type": "piercing"}]',
    NULL,
    'Brutish giants that hunt other creatures for food, often working as mercenaries for more intelligent evil creatures.',
    'Monster Manual'
),

-- CR 3-4 creatures
(
    'Minotaur', 'Large', 'monstrosity', NULL, 'chaotic evil', 14, 76, '9d10 + 27', 40,
    18, 11, 16, 6, 16, 9,
    3, 700, 17,
    NULL, NULL, NULL,
    ARRAY['Abyssal'],
    '[{"name": "Greataxe", "description": "Melee Weapon Attack: +6 to hit, reach 5 ft., one target. Hit: 17 (2d12 + 4) slashing damage.", "attack_bonus": 6, "damage": "2d12+4", "damage_type": "slashing"}, {"name": "Gore", "description": "Melee Weapon Attack: +6 to hit, reach 5 ft., one target. Hit: 13 (2d8 + 4) piercing damage.", "attack_bonus": 6, "damage": "2d8+4", "damage_type": "piercing"}]',
    '[{"name": "Charge", "description": "If the minotaur moves at least 10 feet straight toward a target and then hits it with a gore attack on the same turn, the target takes an extra 9 (2d8) piercing damage. If the target is a creature, it must succeed on a DC 14 Strength saving throw or be pushed up to 10 feet away and knocked prone."}, {"name": "Labyrinthine Recall", "description": "The minotaur can perfectly recall any path it has traveled."}, {"name": "Reckless", "description": "At the start of its turn, the minotaur can gain advantage on all melee weapon attack rolls during that turn, but attack rolls against it have advantage until the start of its next turn."}]',
    'A monstrous creature with the head of a bull and the body of a human, cursed to wander endless mazes.',
    'Monster Manual'
),
(
    'Basilisk', 'Medium', 'monstrosity', NULL, 'unaligned', 15, 52, '8d8 + 16', 20,
    16, 8, 15, 2, 8, 7,
    3, 700, 9,
    NULL, NULL, NULL,
    NULL,
    '[{"name": "Bite", "description": "Melee Weapon Attack: +5 to hit, reach 5 ft., one target. Hit: 10 (2d6 + 3) piercing damage plus 7 (2d6) poison damage.", "attack_bonus": 5, "damage": "2d6+3", "damage_type": "piercing", "additional_damage": "2d6 poison"}]',
    '[{"name": "Petrifying Gaze", "description": "If a creature starts its turn within 30 feet of the basilisk and the two of them can see each other, the basilisk can force the creature to make a DC 12 Constitution saving throw if the basilisk isn''t incapacitated. On a failed save, the creature magically begins to turn to stone and is restrained. It must repeat the saving throw at the end of its next turn. On a success, the effect ends. On a failure, the creature is petrified until freed by the greater restoration spell or other magic."}]',
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
    '[{"name": "Multiattack", "description": "The dragon makes three attacks: one with its bite and two with its claws."}, {"name": "Bite", "description": "Melee Weapon Attack: +7 to hit, reach 10 ft., one target. Hit: 15 (2d10 + 4) piercing damage plus 4 (1d8) acid damage.", "attack_bonus": 7, "damage": "2d10+4", "damage_type": "piercing", "additional_damage": "1d8 acid"}, {"name": "Claw", "description": "Melee Weapon Attack: +7 to hit, reach 5 ft., one target. Hit: 11 (2d6 + 4) slashing damage.", "attack_bonus": 7, "damage": "2d6+4", "damage_type": "slashing"}, {"name": "Acid Breath", "description": "The dragon exhales acid in a 30-foot line that is 5 feet wide. Each creature in that area must make a DC 14 Dexterity saving throw, taking 49 (11d8) acid damage on a failed save, or half as much damage on a successful one.", "save_dc": 14, "save_ability": "Dexterity", "damage": "11d8", "damage_type": "acid", "area": "30-foot line", "recharge": "5-6"}]',
    '[{"name": "Amphibious", "description": "The dragon can breathe air and water."}]',
    'A malevolent dragon that dwells in swamps and marshes, delighting in corruption and decay.',
    'Monster Manual'
),
(
    'Troll', 'Large', 'giant', NULL, 'chaotic evil', 15, 84, '8d10 + 40', 30,
    18, 13, 20, 7, 9, 7,
    5, 1800, 12,
    NULL, NULL, NULL,
    ARRAY['Giant'],
    '[{"name": "Multiattack", "description": "The troll makes three attacks: one with its bite and two with its claws."}, {"name": "Bite", "description": "Melee Weapon Attack: +7 to hit, reach 5 ft., one target. Hit: 7 (1d6 + 4) piercing damage.", "attack_bonus": 7, "damage": "1d6+4", "damage_type": "piercing"}, {"name": "Claw", "description": "Melee Weapon Attack: +7 to hit, reach 5 ft., one target. Hit: 11 (2d6 + 4) slashing damage.", "attack_bonus": 7, "damage": "2d6+4", "damage_type": "slashing"}]',
    '[{"name": "Keen Smell", "description": "The troll has advantage on Wisdom (Perception) checks that rely on smell."}, {"name": "Regeneration", "description": "The troll regains 10 hit points at the start of its turn. If the troll takes acid or fire damage, this trait doesn''t function at the start of the troll''s next turn. The troll dies only if it starts its turn with 0 hit points and doesn''t regenerate."}]',
    'A fearsome giant with incredible regenerative abilities, capable of regrowing lost limbs.',
    'Monster Manual'
),

-- Spellcaster creatures
(
    'Drow', 'Medium', 'humanoid', 'elf', 'neutral evil', 15, 13, '3d8', 30,
    10, 14, 10, 11, 11, 12,
    0.25, 50, 12,
    NULL, NULL, NULL,
    ARRAY['Elvish', 'Undercommon'],
    '[{"name": "Shortsword", "description": "Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6 + 2) piercing damage.", "attack_bonus": 4, "damage": "1d6+2", "damage_type": "piercing"}, {"name": "Hand Crossbow", "description": "Ranged Weapon Attack: +4 to hit, range 30/120 ft., one target. Hit: 5 (1d6 + 2) piercing damage, and the target must succeed on a DC 13 Constitution saving throw or be poisoned for 1 hour. If the saving throw fails by 5 or more, the target is also unconscious while poisoned in this way.", "attack_bonus": 4, "damage": "1d6+2", "damage_type": "piercing", "special": "DC 13 Constitution save or poisoned"}]',
    '[{"name": "Fey Ancestry", "description": "The drow has advantage on saving throws against being charmed, and magic can''t put the drow to sleep."}, {"name": "Innate Spellcasting", "description": "The drow''s spellcasting ability is Charisma (spell save DC 11). It can innately cast the following spells, requiring no material components: At will: dancing lights. 1/day each: darkness, faerie fire."}, {"name": "Sunlight Sensitivity", "description": "While in sunlight, the drow has disadvantage on attack rolls, as well as on Wisdom (Perception) checks that rely on sight."}]',
    'Dark elves that dwell in the Underdark, known for their cruelty and spider worship.',
    'Monster Manual'
);