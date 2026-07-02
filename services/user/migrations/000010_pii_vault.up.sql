-- Blind index for phone lookup when phone column holds Vault ciphertext.
ALTER TABLE "user".users
    ADD COLUMN IF NOT EXISTS phone_lookup TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_phone_lookup
    ON "user".users (phone_lookup)
    WHERE phone_lookup IS NOT NULL;

-- Encrypted DOB replaces plain DATE at the application layer.
ALTER TABLE "user".profiles
    ADD COLUMN IF NOT EXISTS date_of_birth_encrypted TEXT;