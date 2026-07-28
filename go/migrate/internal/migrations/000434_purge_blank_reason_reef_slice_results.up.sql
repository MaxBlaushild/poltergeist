-- Data cleanup, not a schema change. go/pkg/reef/validate.Validate has never
-- had a code path that produces an empty Rejection.Reason (every one of its
-- six rules always constructs a full sentence — confirmed via git history:
-- only one commit has ever touched validate.go, and it already had every
-- Reason populated). Any reef_slice_results row with status='rejected' and
-- a blank rejection_reason is therefore not something the application ever
-- wrote — most likely a hand-inserted test/fixture row from testing the
-- "rejected" UI state directly against the DB.
--
-- This matters because reef_slice_results is a permanent cache keyed by
-- geometry_hash (R-3.3: "identical inputs must never regenerate or
-- re-slice") — Create() uses ON CONFLICT (geometry_hash) DO NOTHING, so any
-- such row permanently poisons every future attempt at that exact parameter
-- combination, silently discarding the real (correctly-reasoned) result
-- computed on each retry. Deleting it lets the next attempt populate the
-- cache correctly instead of losing the conflict race to stale data forever.
DELETE FROM reef_slice_results
WHERE status = 'rejected'
  AND (rejection_reason IS NULL OR rejection_reason = '');
