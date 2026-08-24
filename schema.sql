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
CREATE TABLE "handles" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	"user_id" uuid NOT NULL,
	"handle" varchar(64) NOT NULL CONSTRAINT "handles_handle_key" UNIQUE,
	"status" varchar(32) DEFAULT 'active' NOT NULL,
	"is_primary" boolean DEFAULT false,
	"created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX "handles_handle_key" ON "handles" ("handle");
CREATE UNIQUE INDEX "handles_pkey" ON "handles" ("id");
CREATE INDEX "idx_handles_handle" ON "handles" ("handle");
CREATE INDEX "idx_handles_user_id" ON "handles" ("user_id");


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

-- ============================================================================
-- Indexes for Performance Optimization & Lookups
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_handles_name ON handles(name);
CREATE INDEX IF NOT EXISTS idx_handles_user_id ON handles(user_id);

CREATE INDEX IF NOT EXISTS idx_wallets_address ON wallets(address);
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);

CREATE INDEX IF NOT EXISTS idx_oauth_sessions_user_id ON oauth_sessions(user_id);


CREATE TABLE "oauth_clients" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	"client_id" varchar(128) UNIQUE NOT NULL,
	"client_secret" varchar(256) NOT NULL,
	"redirect_uri" text NOT NULL,
	"name" varchar(128) NOT NULL
);

CREATE TABLE "oauth_codes" (
	"code" varchar(128) PRIMARY KEY,
	"client_id" varchar(128) NOT NULL,
	"user_id" uuid NOT NULL,
	"expires_at" timestamp with time zone NOT NULL
);
