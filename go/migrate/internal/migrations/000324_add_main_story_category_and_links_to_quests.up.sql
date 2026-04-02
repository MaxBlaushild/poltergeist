ALTER TABLE quest_archetypes
ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT 'side';

ALTER TABLE quest_archetypes
DROP CONSTRAINT IF EXISTS quest_archetypes_category_check;

ALTER TABLE quest_archetypes
ADD CONSTRAINT quest_archetypes_category_check
CHECK (category IN ('side', 'main_story'));

ALTER TABLE quests
ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT 'side',
ADD COLUMN IF NOT EXISTS main_story_previous_quest_id UUID REFERENCES quests(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS main_story_next_quest_id UUID REFERENCES quests(id) ON DELETE SET NULL;

ALTER TABLE quests
DROP CONSTRAINT IF EXISTS quests_category_check;

ALTER TABLE quests
ADD CONSTRAINT quests_category_check
CHECK (category IN ('side', 'main_story'));

UPDATE quest_archetypes
SET category = 'main_story'
WHERE EXISTS (
    SELECT 1
    FROM jsonb_array_elements_text(COALESCE(internal_tags, '[]'::jsonb)) AS tag(value)
    WHERE LOWER(TRIM(tag.value)) = 'main_story'
);

UPDATE quests
SET category = 'main_story'
WHERE quest_archetype_id IN (
    SELECT id FROM quest_archetypes WHERE category = 'main_story'
);
