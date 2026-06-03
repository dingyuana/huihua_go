-- 022_seed_data.sql: Initial seed data for development
-- Run after all schema migrations complete

-- Insert default tenant
INSERT INTO tenants (id, name, plan)
VALUES ('00000000-0000-0000-0000-000000000001', '慧话财务', 'enterprise')
ON CONFLICT (id) DO NOTHING;

-- Insert admin user (password: admin123, bcrypt hash)
-- Hash generated with: bcrypt hash of 'admin123'
INSERT INTO users (id, tenant_id, username, password_hash, role, is_active)
VALUES (
    '00000000-0000-0000-0000-000000000101',
    '00000000-0000-0000-0000-000000000001',
    'admin',
    '$2b$12$SjdQr6oJ1CU8SXdwrmoLtObslvYZXndB08Ttl.xdGTpzaJtqA9V7W',
    'admin',
    TRUE
)
ON CONFLICT (tenant_id, username) DO NOTHING;
