-- ============================================================================
-- NID Backend - Complete Database Schema Setup
-- Database: PostgreSQL 12+
-- ============================================================================

-- Enable UUID generation extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. Users Table
-- Core table representing system user accounts
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Handles Table
-- Manages .nid identity handles mapped to user accounts
CREATE TABLE IF NOT EXISTS handles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    handle VARCHAR(64) NOT NULL CONSTRAINT handles_handle_key UNIQUE,
    status VARCHAR(32) DEFAULT 'active' NOT NULL,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Wallets Table
-- Stores multi-chain cryptographic wallet addresses (EVM, Solana, etc.) linked to users
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chain VARCHAR(32) NOT NULL,
    network VARCHAR(32) NOT NULL,
    address VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'verified',
    linked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. OAuth Sessions Table
-- Tracks active third-party application login sessions and tokens
CREATE TABLE IF NOT EXISTS oauth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. OAuth Clients Table (OIDC Third-Party Apps)
CREATE TABLE IF NOT EXISTS oauth_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(128) UNIQUE NOT NULL,
    client_secret VARCHAR(256) NOT NULL,
    redirect_uri TEXT NOT NULL,
    name VARCHAR(128) NOT NULL
);

-- 6. OAuth Codes Table (Temporary Authorization Codes)
CREATE TABLE IF NOT EXISTS oauth_codes (
    code VARCHAR(128) PRIMARY KEY,
    client_id VARCHAR(128) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- ============================================================================
-- Indexes for Performance Optimization & Lookups
-- ============================================================================

CREATE UNIQUE INDEX IF NOT EXISTS handles_handle_key ON handles(handle);
CREATE INDEX IF NOT EXISTS idx_handles_handle ON handles(handle);
CREATE INDEX IF NOT EXISTS idx_handles_user_id ON handles(user_id);

CREATE INDEX IF NOT EXISTS idx_wallets_address ON wallets(address);
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);

CREATE INDEX IF NOT EXISTS idx_oauth_sessions_user_id ON oauth_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_codes_client_id ON oauth_codes(client_id);
