-- Enforce that price must be positive
ALTER TABLE products ADD CONSTRAINT price_positive CHECK (price > 0);

-- Enforce that name cannot be empty or just whitespace
ALTER TABLE products ADD CONSTRAINT name_not_empty CHECK (TRIM(name) <> '');
