-- Safety net for score data: every game reset snapshots the House Favor and
-- Blood Token ledgers into these append-only archive tables before deleting the
-- live rows, so a reset (accidental or deliberate) is always recoverable.
CREATE TABLE IF NOT EXISTS vampire_house_favor_ledger_archive (
    LIKE vampire_house_favor_ledger INCLUDING DEFAULTS,
    archived_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS vampire_blood_token_log_archive (
    LIKE vampire_blood_token_log INCLUDING DEFAULTS,
    archived_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
