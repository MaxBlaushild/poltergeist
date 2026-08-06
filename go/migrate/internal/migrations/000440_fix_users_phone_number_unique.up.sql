-- users.PhoneNumber is a plain (non-pointer) string in models.User, used
-- across every domain in this monorepo, so email/Google-only accounts
-- (reef-site) get "" rather than NULL for phone_number on insert. A plain
-- UNIQUE constraint treats "" as a real, colliding value — only the very
-- first non-phone account could ever be created. Swap it for a partial
-- unique index that only enforces uniqueness for real phone numbers.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_number_key;
CREATE UNIQUE INDEX IF NOT EXISTS users_phone_number_key ON users (phone_number) WHERE phone_number <> '';
