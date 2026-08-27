DROP TABLE IF EXISTS oauth_sessions CASCADE;

CREATE TABLE oauth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL,
    client_id VARCHAR(128) NOT NULL,
    client_name VARCHAR(128),

    status VARCHAR(32) NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,

    CONSTRAINT oauth_sessions_status_check
        CHECK (
            status IN ('active', 'revoked', 'expired')
        ),

    CONSTRAINT oauth_sessions_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_oauth_sessions_client_id
    ON oauth_sessions (client_id);

CREATE INDEX idx_oauth_sessions_status
    ON oauth_sessions (status);

CREATE INDEX idx_oauth_sessions_user_id
    ON oauth_sessions (user_id);

CREATE INDEX idx_oauth_sessions_user_client
    ON oauth_sessions (user_id, client_id);
