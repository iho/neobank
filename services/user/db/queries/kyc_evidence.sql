-- name: InsertKYCSubmission :exec
INSERT INTO "user".kyc_submissions (
    id, kyc_case_id, user_id, document_type, document_number,
    provider, provider_reference, provider_response,
    screening_decision, screening_reason, correlation_id
) VALUES (
    @id, @kyc_case_id, @user_id, @document_type, @document_number,
    @provider, @provider_reference, @provider_response,
    @screening_decision, @screening_reason, @correlation_id
);

-- name: InsertScreeningCheck :exec
INSERT INTO "user".screening_checks (
    id, check_type, subject_user_id, entity_type, entity_id,
    decision, reason_code, provider, provider_reference, raw_response, correlation_id
) VALUES (
    @id, @check_type, @subject_user_id, @entity_type, @entity_id,
    @decision, @reason_code, @provider, @provider_reference, @raw_response, @correlation_id
);