-- Rollback soft-delete column
DROP INDEX IF EXISTS idx_products_is_active;
ALTER TABLE products DROP COLUMN IF EXISTS is_active;
