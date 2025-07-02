CREATE TABLE monsters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Basic Information
    name VARCHAR(255) NOT NULL UNIQUE,
    size VARCHAR(50) NOT NULL DEFAULT 'Medium', -- Tiny, Small, Medium, Large, Huge, Gargantuan
    type VARCHAR(100) NOT NULL, -- beast, humanoid, dragon, etc.
    subtype VARCHAR(100), -- elf, dwarf, red dragon, etc.
    alignment VARCHAR(100) NOT NULL DEFAULT 'unaligned',
    
    -- Core Stats
    armor_class INTEGER NOT NULL DEFAULT 10,
    hit_points INTEGER NOT NULL DEFAULT 1,
    hit_dice VARCHAR(50), -- e.g., "2d8 + 2"
    speed INTEGER NOT NULL DEFAULT 30, -- base walking speed in feet
    speed_modifiers JSONB, -- {"fly": 60, "swim": 30, "burrow": 20}
    
    -- Ability Scores
    strength INTEGER NOT NULL DEFAULT 10,
    dexterity INTEGER NOT NULL DEFAULT 10,
    constitution INTEGER NOT NULL DEFAULT 10,
    intelligence INTEGER NOT NULL DEFAULT 10,
    wisdom INTEGER NOT NULL DEFAULT 10,
    charisma INTEGER NOT NULL DEFAULT 10,
    
    -- Derived Stats
    proficiency_bonus INTEGER NOT NULL DEFAULT 2,
    challenge_rating DECIMAL(4,2) NOT NULL DEFAULT 0, -- 0, 0.125, 0.25, 0.5, 1, 2, etc.
    experience_points INTEGER NOT NULL DEFAULT 0,
    
    -- Skills and Saves
    saving_throw_proficiencies TEXT[], -- ["Dexterity", "Constitution"]
    skill_proficiencies JSONB, -- {"Perception": 4, "Stealth": 6}
    
    -- Resistances and Immunities
    damage_vulnerabilities TEXT[],
    damage_resistances TEXT[],
    damage_immunities TEXT[],
    condition_immunities TEXT[],
    
    -- Senses
    blindsight INTEGER DEFAULT 0,
    darkvision INTEGER DEFAULT 0,
    tremorsense INTEGER DEFAULT 0,
    truesight INTEGER DEFAULT 0,
    passive_perception INTEGER NOT NULL DEFAULT 10,
    
    -- Languages
    languages TEXT[], -- ["Common", "Draconic"]
    
    -- Special Abilities, Actions, etc.
    special_abilities JSONB, -- [{"name": "Pack Tactics", "description": "..."}]
    actions JSONB, -- [{"name": "Bite", "description": "...", "attack_bonus": 4, "damage": "1d6+2"}]
    legendary_actions JSONB, -- for legendary creatures
    legendary_actions_per_turn INTEGER DEFAULT 0,
    reactions JSONB, -- [{"name": "Opportunity Attack", "description": "..."}]
    
    -- Visual and Flavor
    image_url TEXT,
    description TEXT,
    flavor_text TEXT,
    environment TEXT, -- "Forest, Grassland"
    
    -- Meta
    source VARCHAR(100) DEFAULT 'Custom', -- "Monster Manual", "Custom", etc.
    active BOOLEAN DEFAULT TRUE
);

-- Create indexes for common queries
CREATE INDEX idx_monsters_challenge_rating ON monsters(challenge_rating);
CREATE INDEX idx_monsters_type ON monsters(type);
CREATE INDEX idx_monsters_size ON monsters(size);
CREATE INDEX idx_monsters_active ON monsters(active);

-- Seed with some classic D&D monsters
INSERT INTO monsters (
    name, size, type, alignment, armor_class, hit_points, hit_dice, speed,
    strength, dexterity, constitution, intelligence, wisdom, charisma,
    challenge_rating, experience_points, passive_perception,
    actions, description, source
) VALUES 
(
    'Goblin', 'Small', 'humanoid', 'neutral evil', 15, 7, '2d6', 30,
    8, 14, 10, 10, 8, 8,
    0.25, 50, 9,
    '[{"name": "Scimitar", "description": "Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6 + 2) slashing damage.", "attack_bonus": 4, "damage": "1d6+2", "damage_type": "slashing"}, {"name": "Shortbow", "description": "Ranged Weapon Attack: +4 to hit, range 80/320 ft., one target. Hit: 5 (1d6 + 2) piercing damage.", "attack_bonus": 4, "damage": "1d6+2", "damage_type": "piercing"}]',
    'A small, black-hearted humanoid that gathers in overwhelming numbers to overrun enemies.',
    'Monster Manual'
),
(
    'Orc', 'Medium', 'humanoid', 'chaotic evil', 13, 15, '2d8 + 2', 30,
    16, 12, 16, 7, 11, 10,
    0.5, 100, 10,
    '[{"name": "Greataxe", "description": "Melee Weapon Attack: +5 to hit, reach 5 ft., one target. Hit: 9 (1d12 + 3) slashing damage.", "attack_bonus": 5, "damage": "1d12+3", "damage_type": "slashing"}, {"name": "Javelin", "description": "Melee or Ranged Weapon Attack: +5 to hit, reach 5 ft. or range 30/120 ft., one target. Hit: 6 (1d6 + 3) piercing damage.", "attack_bonus": 5, "damage": "1d6+3", "damage_type": "piercing"}]',
    'Savage raiders and pillagers with stooped postures, low foreheads, and piggish faces.',
    'Monster Manual'
),
(
    'Wolf', 'Medium', 'beast', 'unaligned', 13, 11, '2d8 + 2', 40,
    12, 15, 12, 3, 12, 6,
    0.25, 50, 13,
    '[{"name": "Bite", "description": "Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 7 (2d4 + 2) piercing damage. If the target is a creature, it must succeed on a DC 11 Strength saving throw or be knocked prone.", "attack_bonus": 4, "damage": "2d4+2", "damage_type": "piercing", "special": "DC 11 Strength save or prone"}]',
    'A keen-eyed predator that roams forests and grasslands in packs.',
    'Monster Manual'
),
(
    'Adult Red Dragon', 'Huge', 'dragon', 'chaotic evil', 19, 256, '19d12 + 133', 40,
    27, 10, 25, 16, 13, 21,
    17, 18000, 26,
    '[{"name": "Multiattack", "description": "The dragon can use its Frightful Presence. It then makes three attacks: one with its bite and two with its claws."}, {"name": "Bite", "description": "Melee Weapon Attack: +17 to hit, reach 10 ft., one target. Hit: 21 (2d10 + 10) piercing damage plus 14 (4d6) fire damage.", "attack_bonus": 17, "damage": "2d10+10", "damage_type": "piercing", "additional_damage": "4d6 fire"}, {"name": "Fire Breath", "description": "The dragon exhales fire in a 60-foot cone. Each creature in that area must make a DC 21 Dexterity saving throw, taking 63 (18d6) fire damage on a failed save, or half as much damage on a successful one.", "save_dc": 21, "save_ability": "Dexterity", "damage": "18d6", "damage_type": "fire", "area": "60-foot cone"}]',
    'The most covetous and arrogant of all chromatic dragons, red dragons are tyrants of the highest order.',
    'Monster Manual'
),
(
    'Skeleton', 'Medium', 'undead', 'lawful evil', 13, 13, '2d8 + 4', 30,
    10, 14, 15, 6, 8, 5,
    0.25, 50, 9,
    '[{"name": "Shortsword", "description": "Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6 + 2) piercing damage.", "attack_bonus": 4, "damage": "1d6+2", "damage_type": "piercing"}, {"name": "Shortbow", "description": "Ranged Weapon Attack: +4 to hit, range 80/320 ft., one target. Hit: 5 (1d6 + 2) piercing damage.", "attack_bonus": 4, "damage": "1d6+2", "damage_type": "piercing"}]',
    'Animated bones of a dead creature, held together by dark magic.',
    'Monster Manual'
);