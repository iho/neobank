ALTER TABLE card.cards
    DROP COLUMN IF EXISTS daily_limit,
    DROP COLUMN IF EXISTS online_only;