-- Free-text personality/trait tags on a character ("musical", "gambler",
-- "aggressive", "risk taker", ...) — shared content, edited by super users
-- alongside the rest of a character's bio. Used by the Invites tab to
-- filter/search the "who is this invite for" picker. JSONB (a string
-- array, e.g. ["musical","gambler"]) to match this codebase's existing
-- StringArray convention (see models.StringArray) rather than a native
-- Postgres text[]. Nullable, not NOT NULL: cmd/seed's UpsertCharacter
-- creates new characters without setting Tags, and StringArray.Value()
-- sends a Go nil slice through as SQL NULL rather than '[]' — the app
-- layer treats null and [] the same either way.
ALTER TABLE vampire_characters ADD COLUMN IF NOT EXISTS tags JSONB DEFAULT '[]'::jsonb;
CREATE INDEX IF NOT EXISTS idx_vampire_characters_tags ON vampire_characters USING GIN (tags);
