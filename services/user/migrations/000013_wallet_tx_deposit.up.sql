ALTER TABLE "user".wallet_transactions
    DROP CONSTRAINT IF EXISTS user_wallet_tx_type_check;

ALTER TABLE "user".wallet_transactions
    ADD CONSTRAINT user_wallet_tx_type_check
        CHECK (tx_type IN ('p2p_out', 'p2p_in', 'card_hold', 'card_purchase', 'deposit'));