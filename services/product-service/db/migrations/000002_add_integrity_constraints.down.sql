-- Rollback integrity constraints
ALTER TABLE products DROP CONSTRAINT IF EXISTS price_positive;
ALTER TABLE products DROP CONSTRAINT IF EXISTS name_not_empty;
