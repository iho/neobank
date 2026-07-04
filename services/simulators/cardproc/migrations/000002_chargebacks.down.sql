DROP TABLE IF EXISTS cardproc.chargebacks;
ALTER TABLE cardproc.transactions DROP COLUMN IF EXISTS expired_at;
