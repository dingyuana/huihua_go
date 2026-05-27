-- 009_app_user.sql: Create non-superuser application role
-- The default POSTGRES_USER (huihua) is a superuser and bypasses RLS.
-- Application must connect as huihua_app to enforce tenant isolation.

-- Create application role (non-superuser, no bypass RLS)
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'huihua_app') THEN
        CREATE ROLE huihua_app WITH LOGIN PASSWORD 'hfpwd_app' NOBYPASSRLS;
    END IF;
END
$$;

-- Grant usage on schema
GRANT USAGE ON SCHEMA public TO huihua_app;

-- Grant SELECT, INSERT, UPDATE, DELETE on all tables
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO huihua_app;

-- Grant usage on all sequences
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO huihua_app;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO huihua_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO huihua_app;

-- Revoke INSERT on audit_logs from huihua_app (only via trigger/procedure)
-- Actually: keep INSERT, but REVOKE UPDATE/DELETE already done in 007_audit.sql
