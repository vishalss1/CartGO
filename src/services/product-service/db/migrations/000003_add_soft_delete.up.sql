-- migration: 000003_add_soft_delete.up.sql
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE;

-- INDEX for filtering active products
CREATE INDEX IF NOT EXISTS idx_products_is_active ON products(is_active);
