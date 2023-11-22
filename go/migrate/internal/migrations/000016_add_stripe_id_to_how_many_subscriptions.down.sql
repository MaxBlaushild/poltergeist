BEGIN;

DROP INDEX idx_stripe_id;
ALTER TABLE how_many_subscriptions DROP COLUMN stripe_id;

COMMIT;