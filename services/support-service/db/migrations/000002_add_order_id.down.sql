-- migration: 000002_add_order_id.down.sql
ALTER TABLE tickets DROP COLUMN IF EXISTS order_id;
