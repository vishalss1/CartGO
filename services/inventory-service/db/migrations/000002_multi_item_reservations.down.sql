-- Restore the single order_id primary key
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS reservations_pkey;
ALTER TABLE reservations ADD PRIMARY KEY (order_id);
