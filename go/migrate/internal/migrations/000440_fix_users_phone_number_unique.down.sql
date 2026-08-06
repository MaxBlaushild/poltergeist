DROP INDEX IF EXISTS users_phone_number_key;
ALTER TABLE users ADD CONSTRAINT users_phone_number_key UNIQUE (phone_number);
