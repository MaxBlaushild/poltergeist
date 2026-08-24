-- The (character_id, ordinal) unique constraints on vampire_secrets and
-- vampire_missions date back to 000407, before either table was
-- mystery-scoped (000461 secrets, 000464 missions). They were never
-- loosened when mystery_id was added, so a character cast in more than one
-- mystery/subplot (e.g. the main mystery plus a sub-plot) can't have a
-- "secret #1" in both — the second insert hits "duplicate key value
-- violates unique constraint vampire_secrets_character_id_ordinal_key",
-- even though CreateSecretForCharacterMystery/ReplaceSecretsForCharacterAndMystery
-- (and their mission equivalents) already scope ordinal assignment per
-- (character, mystery). Ordinal only ever needs to be unique within one
-- character's slice of one mystery, not across every mystery/subplot a
-- character appears in.
ALTER TABLE vampire_secrets DROP CONSTRAINT IF EXISTS vampire_secrets_character_id_ordinal_key;
ALTER TABLE vampire_secrets
  ADD CONSTRAINT vampire_secrets_character_id_mystery_id_ordinal_key UNIQUE (character_id, mystery_id, ordinal);

ALTER TABLE vampire_missions DROP CONSTRAINT IF EXISTS vampire_missions_character_id_ordinal_key;
ALTER TABLE vampire_missions
  ADD CONSTRAINT vampire_missions_character_id_mystery_id_ordinal_key UNIQUE (character_id, mystery_id, ordinal);
