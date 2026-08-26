# NID

<div align="center">

![NID](https://img.shields.io/badge/NID-Identity%20Layer-7c3aed?style=for-the-badge&logo=ethereum&logoColor=white)
![Go](https://img.shields.io/badge/Go-1.26.5-00ADD8?style=for-the-badge&logo=go)
![React](https://img.shields.io/badge/React-18.3.1-61DAFB?style=for-the-badge&logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-5.5.3-3178C6?style=for-the-badge&logo=typescript)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-12%2B-336791?style=for-the-badge&logo=postgresql)
![OAuth](https://img.shields.io/badge/OAuth-2.0%20%2F%20OIDC-4B5563?style=for-the-badge)

</div>

NID is a wallet-backed digital identity platform that combines handle claiming, public profile resolution, wallet authentication, and OAuth/OIDC capabilities in a single repo. The project includes a Go backend and a React + TypeScript frontend, with PostgreSQL as the persistence layer and wallet-driven login flows using EVM and Solana signatures.

This repository is structured as a full-stack identity product with:

- a public handle layer for `.nid` identity claims and resolution
- wallet-based authentication for users and OAuth consent flows
- protected dashboard functionality for handles, wallets, sessions, and profile data
- OAuth 2.0 / OpenID Connect support for third-party applications
- a modern frontend experience for user onboarding and app consent flows

> Repository status: the codebase contains both the backend and frontend, but it is still a working prototype / active development project. Configuration details, database setup, and deployment requirements are defined in the source itself and should be reviewed before production use.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Repository Structure](#repository-structure)
- [Prerequisites](#prerequisites)
- [Environment Setup](#environment-setup)
- [Backend Setup](#backend-setup)
- [Frontend Setup](#frontend-setup)
- [Running the Project](#running-the-project)
- [API Overview](#api-overview)
- [Authentication and Authorization](#authentication-and-authorization)
- [Database](#database)
- [Frontend Routes](#frontend-routes)
- [Usage Examples](#usage-examples)
- [Screenshots](#screenshots)
- [Deployment Notes](#deployment-notes)
- [Security Considerations](#security-considerations)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

NID is designed around the idea of human-readable identity handles that map to wallet ownership. The backend exposes public identity endpoints and protected user routines, while the frontend offers a landing experience, wallet-based claims, dashboard management, and OAuth consent handling.

The project currently includes:

- a Go HTTP API in `nid-backend/`
- a Vite + React + TypeScript frontend in `nid_frontend/`
- a PostgreSQL schema with user, handle, wallet, OAuth, and social tables
- wallet signature verification for both EVM and Solana
- OIDC discovery, authorization, token exchange, and user info endpoints

The backend is started from `cmd/main.go`, where routes are registered manually with the Go standard library `net/http` `ServeMux`. The frontend uses `react-router-dom` and is bootstrapped with Vite.

---

## Key Features

### Identity and handle management

- Claim and resolve unique `.nid` handles
- Public profile lookup by handle, including `@handle` patterns
- Linked wallet ownership checks tied to handle resolution and login
- Authenticated user dashboard for handles and wallet data

### Wallet authentication

- EVM signature verification via `go-ethereum`
- Solana signature verification via `solana-go`
- wallet-based login flow using signed messages
- protected user session via `nid_token` cookie and bearer token support

### OAuth / OpenID Connect

- OAuth client registration, listing, deletion, and secret rotation
- authorization-code flow with required PKCE (`S256`)
- discovery endpoint and JWKS endpoint for OIDC metadata
- access token issuance and userinfo lookup
- third-party application consent approval flow

### User profile and dashboard

- authenticated profile retrieval
- handle listing and wallet linking
- wallet list CRUD
- social identity management with visibility controls
- session listing and revocation
- dashboard pages for overview, analytics, developers, activity, privacy, security, settings, and more

### Frontend experience

- landing page with handle claiming
- login page and dashboard UI
- OAuth authorization approval page
- docs page and public profile rendering
- support for both MetaMask and Phantom wallet flows

---

## Architecture

The application is split into two major runtime layers in the same repository:

```mermaid
flowchart LR
  A[Browser / Wallet] --> B[React Frontend]
  B --> C[Go Backend API]
  C --> D[PostgreSQL Database]
  C --> E[OIDC / OAuth Providers]
  A --> F[Wallet Signature Verification]
  B --> G[OAuth Redirect / Consent Flow]
```

### Backend architecture

The backend is wired in `nid-backend/cmd/main.go` and follows controller → service → repository → database flow.

```text
HTTP request
  -> CORS middleware + request logger
  -> protected route middleware when required
  -> controller
  -> service
  -> repository
  -> PostgreSQL
```

### Frontend architecture

The frontend mounts a browser router and protects dashboard-only routes via `AuthProvider` and `ProtectedRoute`.

```text
App
  -> AuthProvider
  -> BrowserRouter
  -> Route definitions
  -> page components
  -> API client
  -> NID backend
```

### Runtime responsibilities

- backend: identity logic, wallet verification, OAuth/OIDC, persistence, user session handling
- frontend: wallet detection, login UX, dashboard UI, consent UX, public profile rendering

---

## Tech Stack

| Layer         | Technologies                                                       |
| ------------- | ------------------------------------------------------------------ |
| Backend       | Go 1.26.5, `net/http`, `database/sql`                              |
| Database      | PostgreSQL 12+                                                     |
| Wallet auth   | EVM via `go-ethereum`, Solana via `solana-go`                      |
| Security      | `golang-jwt/jwt/v5`, `golang.org/x/crypto`, HMAC, RSA private keys |
| OAuth / OIDC  | custom OAuth 2.0 / OIDC implementation in Go                       |
| Frontend      | React 18, TypeScript, Vite                                         |
| Styling       | Tailwind CSS                                                       |
| UI motion     | `framer-motion`                                                    |
| Icons         | `lucide-react`, `react-icons`                                      |
| Charts        | `recharts`                                                         |
| Runtime state | React Router, Context API                                          |
| Linting       | ESLint                                                             |

---

## Repository Structure

```text
.
├── README.md
├── nid-backend/
│   ├── .env
│   ├── .gitignore
│   ├── README.md
│   ├── Makefile
│   ├── go.mod
│   ├── cmd/
│   │   └── main.go
│   ├── config/
│   │   ├── database.go
│   │   ├── env.go
│   │   └── logger.go
│   ├── database/
│   │   └── db.go
│   ├── migrations/
│   │   ├── 001_initial_schema.sql
│   │   ├── 002_wallets.sql
│   │   ├── 003_oauth_sessions.sql
│   │   └── 004_passkeys.sql
│   ├── models/
│   │   └── entity.go
│   ├── modules/
│   │   ├── auth/
│   │   ├── handle/
│   │   ├── oidc/
│   │   ├── resolution/
│   │   ├── session/
│   │   ├── social/
│   │   ├── user/
│   │   ├── wallet/
│   │   └── wallet_list/
│   ├── pkg/
│   │   ├── helpers/
│   │   └── middleware/
│   ├── schema.sql
│   ├── api_workflow.md
│   ├── arc.md
│   ├── new.json
│   ├── test_api.json
│   ├── user_workflow.md
│   └── working.md
├── nid_frontend/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   ├── index.html
│   ├── eslint.config.js
│   ├── tsconfig.json
│   ├── tsconfig.app.json
│   ├── tsconfig.node.json
│   ├── public/
│   └── src/
│       ├── App.tsx
│       ├── main.tsx
│       ├── index.css
│       ├── api/
│       ├── components/
│       ├── context/
│       ├── data/
│       ├── pages/
│       └── types/
└── README.md
```

---

## Prerequisites

Before running the project, make sure the following are available:

- Go 1.26.5 or compatible toolchain
- PostgreSQL 12+
- `psql` for schema loading
- Node.js 18+ and npm for the frontend
- a browser with MetaMask or Phantom installed for wallet flows

---

## Environment Setup

### Backend environment variables

The backend uses `godotenv.Load()` at startup and falls back to system environment variables if `.env` is missing. The source explicitly reads the following values:

| Variable                   | Required in practice | Default in code                                                      | Purpose                                  |
| -------------------------- | -------------------- | -------------------------------------------------------------------- | ---------------------------------------- |
| `DATABASE_URL`             | Yes                  | `postgres://postgres:postgres@localhost:5432/nid_db?sslmode=disable` | PostgreSQL connection string             |
| `PORT`                     | No                   | `8080`                                                               | HTTP listen port                         |
| `FRONTEND_URL`             | No                   | `http://localhost:5173`                                              | front-end URL used by config             |
| `NID_INTERNAL_AUTH_SECRET` | No                   | `nid-secret-key`                                                     | HMAC secret used by internal auth tokens |
| `NID_OIDC_PRIVATE_KEY`     | No                   | generated RSA key if not present                                     | private key used to sign OIDC ID tokens  |
| `NID_OIDC_ISSUER`          | No                   | `http://localhost:8081`                                              | OIDC issuer                              |
| `NID_OIDC_KEY_ID`          | No                   | `nid-2026-01`                                                        | OIDC signing key ID                      |

Additional note from the code:

- a `JWT_SECRET` value is present in the repo `.env`, but it is not read by the current application code
- the backend `.env` file includes credentials and a private key and should be treated as sensitive material

### Frontend environment variables

The frontend reads `VITE_*` variables at build/runtime. The code references these values:

| Variable            | Used in code                   | Purpose                                                           |
| ------------------- | ------------------------------ | ----------------------------------------------------------------- |
| `VITE_API_BASE_URL` | `src/api/apiClient.ts`         | base URL for API calls; default is `http://localhost:8081/api/v1` |
| `VITE_NID_BACKEND`  | `src/pages/OAuthAuthorize.tsx` | backend base URL for OAuth authorize flows                        |

The project also includes examples and documentation strings using `VITE_API_BASE`, but the actual active client code in `apiClient.ts` primarily uses `VITE_API_BASE_URL`.

### Example backend `.env`

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/nid_db?sslmode=disable
PORT=8081
FRONTEND_URL=http://localhost:5173
NID_INTERNAL_AUTH_SECRET=replace-with-a-long-random-secret
NID_OIDC_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
NID_OIDC_ISSUER=http://localhost:8081
NID_OIDC_KEY_ID=nid-2026-01
```

> Use a real RSA private key in production and rotate it outside of source control. Do not commit secrets.

---

## Backend Setup

The backend is in `nid-backend/` and is started with Go.

### Install dependencies

```bash
cd nid-backend
go mod download
```

### Run backend

```bash
cd nid-backend
go run cmd/main.go
```

### Build backend

```bash
cd nid-backend
make build
```

### Database setup

The project supplies a `Makefile` with:

```bash
make migrate
```

which runs:

```bash
psql $(DATABASE_URL) -f schema.sql
```

Example local database initialization:

```bash
createdb nid_db
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/nid_db?sslmode=disable'
```

Important notes from the codebase:

- `schema.sql` is the current bootstrap schema, but it contains legacy update statements and placeholders that should be reviewed before applying to a fresh database
- `migrations/002_wallets.sql`, `003_oauth_sessions.sql`, and `004_passkeys.sql` are empty in this repo
- `001_initial_schema.sql` is older and not a reliable migration chain by itself
- passkey support is referenced in migration naming but is not implemented in the current app flow

Use the repo’s `schema.sql` carefully and validate it against your target database before production use.

---

## Frontend Setup

The frontend is under `nid_frontend/` and uses Vite + React + TypeScript.

### Install dependencies

```bash
cd nid_frontend
npm install
```

### Run the frontend

```bash
cd nid_frontend
npm run dev
```

### Build frontend

```bash
cd nid_frontend
npm run build
```

### Type check

```bash
cd nid_frontend
npm run typecheck
```

### Lint

```bash
cd nid_frontend
npm run lint
```

---

## Running the Project

Start the backend first, then the frontend.

### Backend

```bash
cd nid-backend
go run cmd/main.go
```

### Frontend

```bash
cd nid_frontend
npm install
npm run dev
```

The backend defaults to port `8080` in config, while the repo’s `.env` sets `PORT=8081` and the frontend defaults to `http://localhost:8081/api/v1`. In practice, the app commonly uses `8081` for backend local development.

### Health check

```bash
curl http://localhost:8081/health
```

Expected response:

```text
OK
```

---

## API Overview

The backend routes are registered in `nid-backend/cmd/main.go`.

### Public endpoints

| Method | Path                                | Description                                         |
| ------ | ----------------------------------- | --------------------------------------------------- |
| `GET`  | `/health`                           | liveness endpoint                                   |
| `POST` | `/api/v1/auth/login`                | verify a wallet signature and issue a session token |
| `POST` | `/api/v1/auth/logout`               | clear the internal auth cookie                      |
| `GET`  | `/api/v1/resolve?handle=&chain=`    | resolve a handle to a wallet address                |
| `POST` | `/api/v1/handles/claim`             | claim a handle and link to wallet                   |
| `GET`  | `/api/v1/{handle}`                  | public profile lookup by handle                     |
| `GET`  | `/api/v1/social/public`             | public social profile list endpoint                 |
| `GET`  | `/oauth/authorize`                  | OAuth authorization request entrypoint              |
| `POST` | `/oauth/authorize/approve`          | approve the authorization request                   |
| `GET`  | `/oauth/client-info`                | fetch OAuth client info                             |
| `POST` | `/oauth/token`                      | exchange auth code for tokens                       |
| `GET`  | `/oauth/userinfo`                   | return authorized user info                         |
| `GET`  | `/.well-known/openid-configuration` | OIDC discovery document                             |
| `GET`  | `/.well-known/jwks.json`            | JSON Web Key Set                                    |

### Protected endpoints

| Method   | Path                                       | Description                         |
| -------- | ------------------------------------------ | ----------------------------------- |
| `POST`   | `/api/v1/oauth/register`                   | register an OAuth client            |
| `GET`    | `/api/v1/oauth/clients`                    | list the user’s OAuth clients       |
| `DELETE` | `/api/v1/oauth/clients/{id}`               | delete an OAuth client              |
| `POST`   | `/api/v1/oauth/clients/{id}/rotate-secret` | rotate client secret                |
| `GET`    | `/api/v1/handles`                          | fetch authenticated user handles    |
| `POST`   | `/api/v1/wallet-list`                      | create wallet-list entry            |
| `GET`    | `/api/v1/wallet-list`                      | list wallet-list entries            |
| `GET`    | `/api/v1/wallet-list/{id}`                 | get wallet-list item                |
| `PUT`    | `/api/v1/wallet-list/{id}`                 | update wallet-list item             |
| `DELETE` | `/api/v1/wallet-list/{id}`                 | delete wallet-list item             |
| `POST`   | `/api/v1/wallets/link`                     | link a wallet                       |
| `GET`    | `/api/v1/sessions`                         | list sessions                       |
| `POST`   | `/api/v1/sessions/revoke`                  | revoke session                      |
| `GET`    | `/api/v1/user/profile`                     | read the authenticated user profile |
| `GET`    | `/api/v1/user/dashboard`                   | dashboard summary                   |
| `GET`    | `/api/v1/auth/me`                          | current authenticated user          |
| `GET`    | `/api/v1/social`                           | list social identities              |
| `GET`    | `/api/v1/social/{id}`                      | read a social identity              |
| `POST`   | `/api/v1/social`                           | create a social identity            |
| `PUT`    | `/api/v1/social/{id}`                      | update social identity              |
| `PATCH`  | `/api/v1/social/{id}/visibility`           | toggle visibility                   |
| `DELETE` | `/api/v1/social/{id}`                      | delete social identity              |

### Example API calls

#### Claim a handle

```bash
curl -X POST http://localhost:8081/api/v1/handles/claim \
  -H 'Content-Type: application/json' \
  -d '{
    "handle": "alice",
    "address": "0x1234567890abcdef1234567890abcdef12345678",
    "chain": "evm",
    "signature": "<wallet-signature>",
    "message": "Claim alice.nid"
  }'
```

#### Log in

```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "handle": "alice",
    "address": "0x1234567890abcdef1234567890abcdef12345678",
    "chain": "evm",
    "signature": "<wallet-signature>",
    "message": "Sign in to NID"
  }'
```

#### Resolve a handle

```bash
curl 'http://localhost:8081/api/v1/resolve?handle=alice&chain=evm'
```

---

## Authentication and Authorization

NID has two related but separate auth flows: its own internal wallet session and OAuth/OIDC authorization for third-party apps.

### 1. Internal wallet login (NID app auth)

The internal login flow checks wallet signatures and links the wallet to the handle owner:

1. browser detects wallet (MetaMask / Phantom)
2. user signs a message
3. frontend sends the wallet payload to `/api/v1/auth/login`
4. backend verifies the signature, checks the wallet and handle, and issues a session token
5. backend sets an `HttpOnly` `nid_token` cookie and supports bearer-token auth

The code comments indicate the cookie is:

- `HttpOnly`
- `SameSite=Lax`
- non-secure in development
- valid for seven days in the current implementation

### 2. OAuth / OIDC authorization

The app supports OAuth 2.0 authorization code flow with PKCE.

Required elements:

- `response_type=code`
- `scope` must include `openid`
- `nonce` is required
- `code_challenge_method` must be `S256`
- `code_challenge` is required

The backend stores authorization code hashes, access token hashes, and session metadata in PostgreSQL. The code comments and schema state:

- authorization codes are single-use and expire after 60 seconds in the current implementation
- access tokens expire after one hour
- ID tokens are RS256 JWTs
- `.well-known/openid-configuration` and `jwks.json` are exposed for OIDC metadata

The frontend OAuth flow is implemented in `src/pages/OAuthAuthorize.tsx` and authenticates the user with a wallet before sending the approval request to the backend.

---

## Database

The database layer is represented by `nid-backend/schema.sql` and `config/database.go`.

### Main tables

- `users`
- `handles`
- `oauth_clients`
- `oauth_codes`
- `oauth_access_tokens`
- `oauth_sessions`
- `oidc_signing_keys`
- `social_identities`
- `wallet_list`

The schema also includes indexes and constraints for:

- unique handle resolution
- OAuth client registration and tokens
- OIDC public key metadata
- social identity normalization and visibility
- wallet list uniqueness by user + chain + network + address

### Data model notes

The current schema uses PostgreSQL UUIDs and stores hashed OAuth secrets/tokens instead of plaintext values in the primary tables. The codebase indicates that the private OIDC signing key should remain outside the database in environment variables or a secrets manager.

---

## Frontend Routes

The frontend is routed with `react-router-dom` in `nid_frontend/src/App.tsx`.

| Route                         | Purpose                     |
| ----------------------------- | --------------------------- |
| `/`                           | landing page / handle claim |
| `/docs`                       | documentation page          |
| `/login`                      | NID login page              |
| `/oauth/authorize`            | OAuth consent approval page |
| `/dashboard`                  | dashboard home              |
| `/dashboard/handles`          | handle management           |
| `/dashboard/wallets`          | wallet management           |
| `/dashboard/sessions`         | session view                |
| `/dashboard/sdk`              | SDK demo                    |
| `/dashboard/passkeys`         | passkeys area               |
| `/dashboard/privacy`          | privacy settings            |
| `/dashboard/payment-routing`  | payment routing             |
| `/dashboard/social-directory` | social directory            |
| `/dashboard/analytics`        | analytics                   |
| `/dashboard/developers`       | developer section           |
| `/dashboard/security`         | security section            |
| `/dashboard/activity`         | activity logs               |
| `/dashboard/settings`         | settings                    |
| `/:handle`                    | public user profile page    |

---

## Usage Examples

### Claim a handle from the frontend

The landing page detects an installed wallet, requests access, signs a message such as `Claim alice.nid`, and posts the result to `/api/v1/handles/claim`.

### Log in through the frontend

The login and OAuth authorize flows both use wallet signatures to authenticate the user and establish the NID session cookie.

### Access protected profile data

```bash
curl http://localhost:8081/api/v1/user/profile \
  -H 'Authorization: Bearer <token>'
```

or with the cookie from the browser session:

```bash
curl http://localhost:8081/api/v1/user/profile \
  -b 'nid_token=<cookie-value>'
```

---

## Screenshots

> Screenshot placeholders for GitHub presentation.

### Landing page

![Landing page placeholder](https://via.placeholder.com/1200x700?text=Landing+Page)

### Dashboard overview

![Dashboard placeholder](https://via.placeholder.com/1200x700?text=Dashboard+Overview)

### OAuth approval flow

![OAuth consent placeholder](https://via.placeholder.com/1200x700?text=OAuth+Consent+Flow)

---

## Deployment Notes

There are no deployment manifests, Dockerfiles, CI configs, or Kubernetes files in this repository as provided. The code is designed to run as a standard Go API plus a static Vite frontend.

### Recommended deployment pattern

- deploy backend as a Go service with environment variables loaded from a secret manager
- deploy frontend as a static web app or reverse-proxy behind a CDN
- use TLS and a trusted domain for the OIDC issuer and OAuth redirect URIs
- secure the database connection string and private OIDC key with real secret storage

### Important implementation caveat

The repository is not yet a hardened production deployment blueprint. The app includes functional wallet auth and OAuth logic, but the code comments and schema notes explicitly warn about development-oriented defaults and the need for a stronger production setup.

---

## Security Considerations

The repository already contains several important security notes in code comments and configuration:

- OAuth client secrets are hashed before storage
- authorization codes and access tokens are stored as hashes, not raw secrets
- the backend supports `HttpOnly` authenticated cookies
- `NID_OIDC_PRIVATE_KEY` is expected to be stored securely and not in plaintext source control
- `NID_INTERNAL_AUTH_SECRET` should be unique and strong
- PKCE is mandatory for OAuth authorization requests

Additional production hardening recommendations:

- rotate the existing backend `.env` secrets before public deployment
- enable HTTPS everywhere
- validate redirect URIs strictly
- add token expiry and revocation auditing
- review session replay protections and refresh flow
- ensure database access is restricted and encrypted in transit
- configure production secrets via environment injection or a vault service

---

## Contributing

Contributions are welcome in principle, but this repository does not appear to include a formal contribution guide or CI workflow at the root level. The project structure is clear enough to contribute in the following areas:

- backend API improvements
- OAuth/OIDC hardening
- wallet verification validation
- frontend UX and route flow polish
- database schema review and migration hygiene
- docs and examples

Before making changes:

1. review the backend code in `nid-backend/cmd/main.go`
2. check the frontend routes in `nid_frontend/src/App.tsx`
3. validate environment and database assumptions
4. avoid committing `.env` secrets or live credentials

---

## License

No license file was found in the repository at the root of the workspace. The project should not be treated as open-source licensed until a license is added.

If this project is intended for public GitHub distribution, add a license file such as MIT, Apache-2.0, or GPL before publishing.

---

## Summary

NID is a wallet-first identity platform with a Go backend and a React frontend, built around handle claiming, wallet-backed authentication, user dashboards, and OAuth/OIDC interoperability. The repository is functionally rich and clearly structured, but it still requires careful production hardening in areas such as database migration hygiene, secret management, and deployment configuration.

This README is intended to document the project as it exists in the current codebase, without overstating completion or production readiness.
