-- name: InsertScreeningCheck :exec
INSERT INTO payment.screening_checks (
    id, check_type, subject_user_id, related_user_id, entity_type, entity_id,
    decision, reason_code, provider, provider_reference, raw_response, correlation_id
) VALUES (
    @id, @check_type, @subject_user_id, @related_user_id, @entity_type, @entity_id,
    @decision, @reason_code, @provider, @provider_reference, @raw_response, @correlation_id
);