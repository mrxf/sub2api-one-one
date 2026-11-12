-- Dedicated account ownership, pre-listed buyout marketplace with daily decay pricing,
-- and API key 1-to-1 / multi-account dedicated binding.

ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS owner_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS display_alias VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS sale_status VARCHAR(20) NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS sale_price NUMERIC(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS daily_price_decay NUMERIC(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS min_sale_price NUMERIC(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS listed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS sold_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS sale_description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS locked_by_order_id BIGINT,
    ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_accounts_owner_user_id
    ON accounts(owner_user_id)
    WHERE deleted_at IS NULL AND owner_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_accounts_sale_status
    ON accounts(sale_status)
    WHERE deleted_at IS NULL AND sale_status <> 'none';

ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS dedicated_account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS dedicated_pool_mode BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_api_keys_dedicated_account_id
    ON api_keys(dedicated_account_id)
    WHERE deleted_at IS NULL AND dedicated_account_id IS NOT NULL;

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS dedicated_account_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_payment_orders_dedicated_account_id
    ON payment_orders(dedicated_account_id)
    WHERE dedicated_account_id IS NOT NULL;
