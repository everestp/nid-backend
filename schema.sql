-- ============================================================================
-- NID Backend - Complete Database Schema
-- Database: PostgreSQL 12+
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- ============================================================================
-- 1. USERS
-- ============================================================================

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- ============================================================================
-- 2. HANDLES
-- ============================================================================

CREATE TABLE IF NOT EXISTS handles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    handle VARCHAR(64) NOT NULL,

    status VARCHAR(32) NOT NULL DEFAULT 'active',

    is_primary BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT handles_handle_key UNIQUE (handle)
);

CREATE INDEX IF NOT EXISTS idx_handles_user_id
    ON handles(user_id);

CREATE INDEX IF NOT EXISTS idx_handles_handle
    ON handles(handle);


-- ============================================================================
-- 3. WALLETS
-- ============================================================================

CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    chain VARCHAR(32) NOT NULL,

    network VARCHAR(32) NOT NULL,

    address VARCHAR(255) NOT NULL UNIQUE,

    status VARCHAR(32) NOT NULL DEFAULT 'verified',

    linked_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_wallets_user_id
    ON wallets(user_id);

CREATE INDEX IF NOT EXISTS idx_wallets_address
    ON wallets(address);


-- ============================================================================
-- 4. OAUTH CLIENTS
--
-- Registered third-party applications.
--
-- Example:
--
-- client_id:
--     nid_client_xxxxxxxxx
--
-- client_secret_hash:
--     bcrypt hash
--
-- redirect_uri:
--     https://myapp.com/auth/callback
--
-- client_type:
--     confidential
--     public
-- ============================================================================

CREATE TABLE IF NOT EXISTS oauth_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    client_id VARCHAR(128) NOT NULL UNIQUE,

    client_secret_hash VARCHAR(255),

    name VARCHAR(128) NOT NULL,

    redirect_uri TEXT NOT NULL,

    client_type VARCHAR(32) NOT NULL DEFAULT 'confidential',

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT oauth_clients_client_type_check
        CHECK (client_type IN ('confidential', 'public'))
);

CREATE INDEX IF NOT EXISTS idx_oauth_clients_client_id
    ON oauth_clients(client_id);


-- ============================================================================
-- 5. OAUTH AUTHORIZATION CODES
--
-- Short-lived, single-use authorization codes.
--
-- IMPORTANT:
-- Store only SHA-256 hash of the actual authorization code.
--
-- The real code is returned to the client.
-- The database stores only its hash.
--
-- PKCE is mandatory.
-- ============================================================================

CREATE TABLE IF NOT EXISTS oauth_codes (
    code_hash VARCHAR(64) PRIMARY KEY,

    client_id VARCHAR(128) NOT NULL,

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    redirect_uri TEXT NOT NULL,

    scope TEXT NOT NULL DEFAULT 'openid',

    nonce TEXT,

    code_challenge VARCHAR(128) NOT NULL,

    code_challenge_method VARCHAR(16) NOT NULL DEFAULT 'S256',

    expires_at TIMESTAMPTZ NOT NULL,

    used_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT oauth_codes_pkce_method_check
        CHECK (code_challenge_method = 'S256')
);

CREATE INDEX IF NOT EXISTS idx_oauth_codes_client_id
    ON oauth_codes(client_id);

CREATE INDEX IF NOT EXISTS idx_oauth_codes_user_id
    ON oauth_codes(user_id);

CREATE INDEX IF NOT EXISTS idx_oauth_codes_expires_at
    ON oauth_codes(expires_at);


-- ============================================================================
-- 6. OAUTH ACCESS TOKENS
--
-- OAuth access tokens used by applications to call /oauth/userinfo.
--
-- Only the SHA-256 hash is stored.
-- ============================================================================

CREATE TABLE IF NOT EXISTS oauth_access_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    token_hash VARCHAR(64) NOT NULL UNIQUE,

    client_id VARCHAR(128) NOT NULL,

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    scope TEXT NOT NULL DEFAULT 'openid',

    expires_at TIMESTAMPTZ NOT NULL,

    revoked_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_oauth_access_tokens_hash
    ON oauth_access_tokens(token_hash);

CREATE INDEX IF NOT EXISTS idx_oauth_access_tokens_client_id
    ON oauth_access_tokens(client_id);

CREATE INDEX IF NOT EXISTS idx_oauth_access_tokens_user_id
    ON oauth_access_tokens(user_id);

CREATE INDEX IF NOT EXISTS idx_oauth_access_tokens_expires_at
    ON oauth_access_tokens(expires_at);


-- ============================================================================
-- 7. OAUTH SESSIONS
--
-- Tracks NID login sessions created through third-party applications.
-- ============================================================================

CREATE TABLE IF NOT EXISTS oauth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    status VARCHAR(32) NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT oauth_sessions_status_check
        CHECK (status IN ('active', 'revoked', 'expired'))
);

CREATE INDEX IF NOT EXISTS idx_oauth_sessions_user_id
    ON oauth_sessions(user_id);

CREATE INDEX IF NOT EXISTS idx_oauth_sessions_status
    ON oauth_sessions(status);


-- ============================================================================
-- 8. OIDC SIGNING KEYS
--
-- Stores metadata for the key used to sign ID tokens.
--
-- IMPORTANT:
-- Do NOT store the private signing key directly in this table unless you
-- properly encrypt it.
--
-- Prefer storing the private key in environment variables / KMS / secrets
-- manager and storing only public JWKS metadata here.
-- ============================================================================

CREATE TABLE IF NOT EXISTS oidc_signing_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    kid VARCHAR(128) NOT NULL UNIQUE,

    algorithm VARCHAR(16) NOT NULL DEFAULT 'RS256',

    key_type VARCHAR(16) NOT NULL DEFAULT 'RSA',

    use VARCHAR(16) NOT NULL DEFAULT 'sig',

    public_key TEXT NOT NULL,

    active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_oidc_signing_keys_active
    ON oidc_signing_keys(active);


-- ============================================================================
-- 9. CLEANUP INDEXES
-- ============================================================================
--
-- Useful for background cleanup jobs.
--

CREATE INDEX IF NOT EXISTS idx_oauth_codes_cleanup
    ON oauth_codes(expires_at);

CREATE INDEX IF NOT EXISTS idx_oauth_access_tokens_cleanup
    ON oauth_access_tokens(expires_at);


-- ============================================================================
-- END
-- ============================================================================
