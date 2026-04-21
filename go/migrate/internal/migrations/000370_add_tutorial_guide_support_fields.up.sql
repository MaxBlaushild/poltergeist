ALTER TABLE tutorial_configs
  ADD COLUMN IF NOT EXISTS guide_support_greeting TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS guide_support_personality TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS guide_support_behavior TEXT NOT NULL DEFAULT '';
