-- Script to initialize multiple databases for the single PostgreSQL instance
CREATE DATABASE IF NOT EXISTS cartgo_user_db;
CREATE DATABASE IF NOT EXISTS cartgo_product_db;
CREATE DATABASE IF NOT EXISTS cartgo_inventory_db;
CREATE DATABASE IF NOT EXISTS cartgo_order_db;
CREATE DATABASE IF NOT EXISTS cartgo_payment_db;
CREATE DATABASE IF NOT EXISTS cartgo_delivery_db;
CREATE DATABASE IF NOT EXISTS cartgo_support_db;
