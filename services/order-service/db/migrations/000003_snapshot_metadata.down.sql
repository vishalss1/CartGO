-- migration: 000003_snapshot_metadata.down.sql
ALTER TABLE order_items DROP COLUMN IF EXISTS product_name;
ALTER TABLE order_items DROP COLUMN IF EXISTS category;
