CREATE TABLE spell_progressions (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  name TEXT NOT NULL,
  ability_type TEXT NOT NULL DEFAULT 'spell'
);

CREATE TABLE spell_progression_spells (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  progression_id UUID NOT NULL REFERENCES spell_progressions(id) ON DELETE CASCADE,
  spell_id UUID NOT NULL REFERENCES spells(id) ON DELETE CASCADE,
  level_band INTEGER NOT NULL
);

CREATE UNIQUE INDEX spell_progression_spells_progression_spell_uidx
  ON spell_progression_spells(progression_id, spell_id);

CREATE UNIQUE INDEX spell_progression_spells_progression_band_uidx
  ON spell_progression_spells(progression_id, level_band);

CREATE UNIQUE INDEX spell_progression_spells_spell_uidx
  ON spell_progression_spells(spell_id);

CREATE INDEX spell_progression_spells_progression_idx
  ON spell_progression_spells(progression_id);
