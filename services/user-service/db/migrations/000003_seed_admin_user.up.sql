-- Seed admin user (password: admin123)
-- Hash generated via bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
VALUES (
    'admin',
    'admin@cartgo.com',
    '$2a$10$7Z8v7.Wz.M6Z9e/f.I.B.Od3J/06v8h/X.SgWq6h.I.B.Od3J/06v8h', 
    'ADMIN',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
) ON CONFLICT (email) DO NOTHING;
