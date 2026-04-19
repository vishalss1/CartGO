-- migration: 000002_add_delivery_address.down.sql
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_address;
