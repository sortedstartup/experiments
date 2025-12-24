-- PostgreSQL complete schema migration

-- Users table
CREATE TABLE IF NOT EXISTS sortedauth_users (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Providers table for OAuth configuration
CREATE TABLE IF NOT EXISTS sortedauth_providers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    oauth_url TEXT NOT NULL,
    token_url TEXT NOT NULL,
    redirecturl TEXT NOT NULL,
    scope TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Accounts table - links users to OAuth providers (1 user can have many accounts)
CREATE TABLE IF NOT EXISTS sortedauth_accounts (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES sortedauth_users(user_id),
    provider TEXT NOT NULL,
    provider_account_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_account_id)
);

-- Tenants table
CREATE TABLE IF NOT EXISTS sortedauth_tenants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    type INTEGER NOT NULL DEFAULT 1, -- 0 = system, 1 = organization, 2 = personal
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL REFERENCES sortedauth_users(user_id)
);

-- Tenant users table - many-to-many relationship between tenants and users
CREATE TABLE IF NOT EXISTS sortedauth_tenant_users (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES sortedauth_tenants(id),
    user_id TEXT NOT NULL REFERENCES sortedauth_users(user_id),
    role TEXT NOT NULL DEFAULT 'user', -- Simple string: 'admin', 'user', etc.
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, user_id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_sortedauth_users_email ON sortedauth_users(email);
CREATE INDEX IF NOT EXISTS idx_sortedauth_accounts_user_id ON sortedauth_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_sortedauth_accounts_provider ON sortedauth_accounts(provider);
CREATE INDEX IF NOT EXISTS idx_sortedauth_tenant_users_tenant_id ON sortedauth_tenant_users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sortedauth_tenant_users_user_id ON sortedauth_tenant_users(user_id);
