-- Dedicated account full buyout (raw account credentials extraction) configuration & status.
-- Once bought_out_at is set, the owner has extracted the plaintext buyout_info and
-- forfeits platform ban-compensation warranty policies.

ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS buyout_price NUMERIC(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS buyout_info TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS bought_out_at TIMESTAMPTZ;
