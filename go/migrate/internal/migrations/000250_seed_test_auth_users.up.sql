INSERT INTO users (id, created_at, updated_at, name, phone_number, active)
VALUES
  ('d8d28ec1-2162-4d87-97d6-c8f8b6e6a801', NOW(), NOW(), 'Demo User', '+14407858475', TRUE),
  ('be105033-b2e6-4d4d-a91c-70037dbcc0f3', NOW(), NOW(), 'Test User 2', '+12025550101', TRUE)
ON CONFLICT (phone_number) DO UPDATE
SET
  name = EXCLUDED.name,
  active = EXCLUDED.active,
  updated_at = NOW();
