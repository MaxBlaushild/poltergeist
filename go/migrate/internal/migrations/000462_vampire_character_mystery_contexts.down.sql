ALTER TABLE vampire_characters ADD COLUMN IF NOT EXISTS post_act1_context TEXT NOT NULL DEFAULT '';

DO $$
DECLARE
  legacy_mystery_id UUID;
BEGIN
  SELECT id INTO legacy_mystery_id FROM vampire_mysteries WHERE name = 'The Crimson Toast' LIMIT 1;
  IF legacy_mystery_id IS NOT NULL THEN
    UPDATE vampire_characters c
    SET post_act1_context = cmc.post_act1_context
    FROM vampire_character_mystery_contexts cmc
    WHERE cmc.character_id = c.id AND cmc.mystery_id = legacy_mystery_id;
  END IF;
END $$;

DROP TABLE IF EXISTS vampire_character_mystery_contexts;
