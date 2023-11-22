BEGIN;

ALTER TABLE how_many_subscriptions ADD COLUMN stripe_id VARCHAR(255);
CREATE INDEX idx_stripe_id ON how_many_subscriptions(stripe_id);

COMMIT;