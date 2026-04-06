-- Remove the single order_id primary key
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS reservations_pkey;

-- Add a composite primary key to allow multiple products per order
ALTER TABLE reservations ADD PRIMARY KEY (order_id, product_id);
