-- migration: 000003_snapshot_metadata.up.sql
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS product_name VARCHAR(255);
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS category VARCHAR(100);

-- Note: existing price_per_unit serves as price_at_purchase
