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


    ALTER TABLE oauth_clients
ADD COLUMN IF NOT EXISTS client_name VARCHAR(255) NOT NULL DEFAULT 'Unknown App',
ADD COLUMN IF NOT EXISTS client_logo VARCHAR(512) DEFAULT '',
ADD COLUMN IF NOT EXISTS client_uri VARCHAR(512) DEFAULT '',
ADD COLUMN IF NOT EXISTS policy_uri VARCHAR(512) DEFAULT '';


-- ============================================================================
-- 4. SOCIAL IDENTITIES
-- ============================================================================
--
-- External social / Web3 / developer identities linked to a NID user.
--
-- Examples:
--   Twitter/X    -> @everest
--   GitHub       -> everest
--   Discord      -> everest
--   Farcaster    -> @everest
--   Telegram     -> everest
--
-- normalized_handle is used for case-insensitive / normalized lookups.
--
-- metadata can store provider-specific information such as:
--   {
--       "url": "https://x.com/everest",
--       "provider_user_id": "123456789"
--   }
--
-- ============================================================================

CREATE TABLE IF NOT EXISTS social_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    platform VARCHAR(32) NOT NULL,

    handle VARCHAR(255) NOT NULL,

    normalized_handle VARCHAR(255) NOT NULL,

    verified BOOLEAN NOT NULL DEFAULT FALSE,

    publicly_visible BOOLEAN NOT NULL DEFAULT FALSE,

    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT social_identities_platform_check
        CHECK (
            platform IN (

                -- Social
                'twitter',
                'instagram',
                'facebook',
                'tiktok',
                'linkedin',
                'threads',
                'bluesky',
                'mastodon',
                'snapchat',
                'pinterest',
                'reddit',

                -- Web3 / Crypto
                'farcaster',
                'lens',
                'warpcast',
                'mirror',
                'zora',
                'opensea',
                'ens',

                -- Developer
                'github',
                'gitlab',
                'bitbucket',
                'stackoverflow',
                'codepen',

                -- Messaging / Community
                'discord',
                'telegram',
                'whatsapp',
                'signal',
                'twitch',

                -- Content
                'youtube',
                'medium',
                'substack',

                -- Personal / Contact
                'email',
                'phone',
                'website'
            )
        ),

    CONSTRAINT social_identities_handle_check
        CHECK (length(trim(handle)) > 0),

    CONSTRAINT social_identities_normalized_handle_check
        CHECK (length(trim(normalized_handle)) > 0),

    CONSTRAINT social_identities_metadata_check
        CHECK (jsonb_typeof(metadata) = 'object'),

    CONSTRAINT social_identities_unique_identity
        UNIQUE (user_id, platform, normalized_handle)
);


-- ============================================================================
-- SOCIAL IDENTITY INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_social_identities_user_id
    ON social_identities(user_id);

CREATE INDEX IF NOT EXISTS idx_social_identities_platform_handle
    ON social_identities(platform, normalized_handle);

CREATE INDEX IF NOT EXISTS idx_social_identities_user_platform
    ON social_identities(user_id, platform);

CREATE INDEX IF NOT EXISTS idx_social_identities_public
    ON social_identities(user_id, publicly_visible)
    WHERE publicly_visible = TRUE;

CREATE INDEX IF NOT EXISTS idx_social_identities_verified
    ON social_identities(user_id, verified)
    WHERE verified = TRUE;

ALTER TABLE oauth_sessions
ADD COLUMN IF NOT EXISTS client_id VARCHAR(128),
  ADD COLUMN IF NOT EXISTS client_name VARCHAR(128),
ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_oauth_sessions_client_id
ON oauth_sessions(client_id);

CREATE INDEX IF NOT EXISTS idx_oauth_sessions_user_client
ON oauth_sessions(user_id, client_id);






ALTER TABLE oauth_access_tokens
ADD COLUMN IF NOT EXISTS session_id UUID
REFERENCES oauth_sessions(id)
ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_oauth_access_tokens_session_id
ON oauth_access_tokens(session_id);

UPDATE oauth_sessions
SET last_used_at = CURRENT_TIMESTAMP
WHERE id = (
    SELECT session_id
    FROM oauth_access_tokens
    WHERE token_hash = $1
);
-- ============================================================================
-- END
-- ============================================================================
