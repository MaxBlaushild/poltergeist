ALTER TABLE tutorial_configs
  DROP COLUMN IF EXISTS guide_support_behavior,
  DROP COLUMN IF EXISTS guide_support_personality,
  DROP COLUMN IF EXISTS guide_support_greeting;
