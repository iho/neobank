-- name: CreateDispute :one
INSERT INTO card.disputes (
    chargeback_id, authorization_id, card_id, user_id, amount, currency, reason,
    provisional_credit_transfer_id
) VALUES (
    @chargeback_id, @authorization_id, @card_id, @user_id, @amount::numeric, @currency, @reason,
    @provisional_credit_transfer_id
)
RETURNING id, chargeback_id, authorization_id, card_id, user_id, amount::text AS amount, currency,
          reason, status, provisional_credit_transfer_id, reversal_transfer_id, created_at, updated_at;

-- name: GetDisputeByChargebackID :one
SELECT id, chargeback_id, authorization_id, card_id, user_id, amount::text AS amount, currency,
       reason, status, provisional_credit_transfer_id, reversal_transfer_id, created_at, updated_at
FROM card.disputes
WHERE chargeback_id = $1;

-- name: MarkDisputeResolved :one
UPDATE card.disputes
SET status = @status, reversal_transfer_id = @reversal_transfer_id, updated_at = now()
WHERE chargeback_id = @chargeback_id
RETURNING id, chargeback_id, authorization_id, card_id, user_id, amount::text AS amount, currency,
          reason, status, provisional_credit_transfer_id, reversal_transfer_id, created_at, updated_at;
