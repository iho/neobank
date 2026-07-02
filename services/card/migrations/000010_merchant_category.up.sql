ALTER TABLE card.authorizations
    ADD COLUMN IF NOT EXISTS merchant_category_code CHAR(4);