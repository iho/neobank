DROP INDEX IF EXISTS "user".idx_user_kyc_cases_vendor_applicant;
ALTER TABLE "user".kyc_cases DROP COLUMN IF EXISTS vendor_applicant_id;
