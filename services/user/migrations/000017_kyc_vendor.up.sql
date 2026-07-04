ALTER TABLE "user".kyc_cases ADD COLUMN IF NOT EXISTS vendor_applicant_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_kyc_cases_vendor_applicant
    ON "user".kyc_cases (vendor_applicant_id)
    WHERE vendor_applicant_id IS NOT NULL;
