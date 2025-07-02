CREATE TABLE dnd_classes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    hit_die INTEGER NOT NULL DEFAULT 8,
    primary_ability VARCHAR(255),
    saving_throw_proficiencies TEXT[],
    skill_options TEXT[],
    equipment_proficiencies TEXT[],
    spell_casting_ability VARCHAR(255),
    is_spellcaster BOOLEAN DEFAULT FALSE,
    active BOOLEAN DEFAULT TRUE
);

-- Seed the table with classic D&D 5e classes
INSERT INTO dnd_classes (name, description, hit_die, primary_ability, saving_throw_proficiencies, skill_options, equipment_proficiencies, spell_casting_ability, is_spellcaster) VALUES
('Fighter', 'Masters of martial combat, skilled with a variety of weapons and armor. Fighters are versatile warriors who excel in both offense and defense.', 10, 'Strength or Dexterity', ARRAY['Strength', 'Constitution'], ARRAY['Acrobatics', 'Animal Handling', 'Athletics', 'History', 'Insight', 'Intimidation', 'Perception', 'Survival'], ARRAY['All armor', 'Shields', 'Simple weapons', 'Martial weapons'], NULL, FALSE),
('Wizard', 'Scholarly magic-users capable of manipulating the structures of reality. Masters of arcane magic through study and preparation.', 6, 'Intelligence', ARRAY['Intelligence', 'Wisdom'], ARRAY['Arcana', 'History', 'Insight', 'Investigation', 'Medicine', 'Religion'], ARRAY['Daggers', 'Darts', 'Slings', 'Quarterstaffs', 'Light crossbows'], 'Intelligence', TRUE),
('Rogue', 'Scoundrels who use stealth and trickery to achieve their goals. Masters of skills, sneak attacks, and avoiding danger.', 8, 'Dexterity', ARRAY['Dexterity', 'Intelligence'], ARRAY['Acrobatics', 'Athletics', 'Deception', 'Insight', 'Intimidation', 'Investigation', 'Perception', 'Performance', 'Persuasion', 'Sleight of Hand', 'Stealth'], ARRAY['Light armor', 'Simple weapons', 'Hand crossbows', 'Longswords', 'Rapiers', 'Shortswords'], NULL, FALSE),
('Cleric', 'Divine magic wielders who serve deities and channel divine power. Healers and support specialists with access to divine magic.', 8, 'Wisdom', ARRAY['Wisdom', 'Charisma'], ARRAY['History', 'Insight', 'Medicine', 'Persuasion', 'Religion'], ARRAY['Light armor', 'Medium armor', 'Shields', 'Simple weapons'], 'Wisdom', TRUE),
('Ranger', 'Warriors of the wilderness who track foes and use nature magic. Skilled hunters and guides with limited spellcasting.', 10, 'Dexterity or Wisdom', ARRAY['Strength', 'Dexterity'], ARRAY['Animal Handling', 'Athletics', 'Insight', 'Investigation', 'Nature', 'Perception', 'Stealth', 'Survival'], ARRAY['Light armor', 'Medium armor', 'Shields', 'Simple weapons', 'Martial weapons'], 'Wisdom', TRUE),
('Barbarian', 'Fierce warriors from the wild who fight with primal ferocity. Masters of rage and physical prowess who shun heavy armor.', 12, 'Strength', ARRAY['Strength', 'Constitution'], ARRAY['Animal Handling', 'Athletics', 'Intimidation', 'Nature', 'Perception', 'Survival'], ARRAY['Light armor', 'Medium armor', 'Shields', 'Simple weapons', 'Martial weapons'], NULL, FALSE),
('Bard', 'Masters of song, speech, and magic who inspire allies and confound enemies. Versatile performers with magical abilities.', 8, 'Charisma', ARRAY['Dexterity', 'Charisma'], ARRAY['Any three'], ARRAY['Light armor', 'Simple weapons', 'Hand crossbows', 'Longswords', 'Rapiers', 'Shortswords'], 'Charisma', TRUE),
('Druid', 'Priests of nature who wield elemental magic and can transform into animals. Guardians of the natural world.', 8, 'Wisdom', ARRAY['Intelligence', 'Wisdom'], ARRAY['Arcana', 'Animal Handling', 'Insight', 'Medicine', 'Nature', 'Perception', 'Religion', 'Survival'], ARRAY['Light armor', 'Medium armor', 'Shields (non-metal)', 'Clubs', 'Daggers', 'Darts', 'Javelins', 'Maces', 'Quarterstaffs', 'Scimitars', 'Sickles', 'Slings', 'Spears'], 'Wisdom', TRUE),
('Monk', 'Masters of martial arts who harness inner power. Agile warriors who use ki energy and are skilled in unarmed combat.', 8, 'Dexterity or Wisdom', ARRAY['Strength', 'Dexterity'], ARRAY['Acrobatics', 'Athletics', 'History', 'Insight', 'Religion', 'Stealth'], ARRAY['Simple weapons', 'Shortswords'], NULL, FALSE),
('Paladin', 'Holy warriors bound by sacred oaths. Fighters with divine magic who protect the innocent and uphold justice.', 10, 'Strength or Charisma', ARRAY['Wisdom', 'Charisma'], ARRAY['Athletics', 'Insight', 'Intimidation', 'Medicine', 'Persuasion', 'Religion'], ARRAY['All armor', 'Shields', 'Simple weapons', 'Martial weapons'], 'Charisma', TRUE),
('Sorcerer', 'Magic wielders born with innate magical ability. Spontaneous spellcasters who shape raw magical energy.', 6, 'Charisma', ARRAY['Constitution', 'Charisma'], ARRAY['Arcana', 'Deception', 'Insight', 'Intimidation', 'Persuasion', 'Religion'], ARRAY['Daggers', 'Darts', 'Slings', 'Quarterstaffs', 'Light crossbows'], 'Charisma', TRUE),
('Warlock', 'Wielders of magic derived from a pact with an extraplanar entity. Spellcasters with unique abilities and limited spell slots.', 8, 'Charisma', ARRAY['Wisdom', 'Charisma'], ARRAY['Arcana', 'Deception', 'History', 'Intimidation', 'Investigation', 'Nature', 'Religion'], ARRAY['Light armor', 'Simple weapons'], 'Charisma', TRUE);