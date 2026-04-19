-- migration: 000001_initial_schema.down.sql
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS tickets;
DROP EXTENSION IF EXISTS "pgcrypto";
