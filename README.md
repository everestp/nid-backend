# NID Backend

NID Backend is a Go HTTP API for passwordless, wallet-backed NID identities. It lets clients claim and resolve `.nid` handles, authenticate with EVM or Solana wallet signatures, manage linked wallets and public profile data, and act as an OAuth 2.0 / OpenID Connect provider for registered applications.

> **Implementation status**
> This repository contains the backend only. There is no frontend, container definition, CI pipeline, deployment manifest, automated test suite, or migration runner. The commands and behavior below are based on the current source code.

<!-- SCREENSHOT: Dashboard -->

## Key Features

- Passwordless wallet-signature login for EVM and Solana addresses.
- Claiming and public resolution of unique NID handles.
- Protected user profile and dashboard endpoints.
- Linking wallets and maintaining an authenticated wallet list.
- Social identity CRUD with platform allowlisting and public visibility controls.
- OAuth 2.0 authorization-code flow with mandatory S256 PKCE.
- OpenID Connect discovery, JWKS, ID tokens, access tokens, and UserInfo.
- OAuth client registration, listing, deletion, and secret rotation.
- PostgreSQL persistence with UUID identifiers and foreign-key relationships.

## Architecture

The application is manually wired in `cmd/main.go`:

```text
HTTP request
	-> CORS middleware and request logger
	-> bearer-token middleware (protected routes only)
	-> controller
	-> service
	-> repository
	-> PostgreSQL
```

The standard library `net/http` `ServeMux` owns routing. Domain code is grouped under `modules/`, with separate controller, DTO, service, and repository packages. Shared authentication helpers and middleware live under `pkg/`.

<!-- SCREENSHOT: Architecture diagram -->

## Tech Stack

- Go `1.26.5`
- PostgreSQL, accessed through `database/sql` and `lib/pq`
- `net/http` and `github.com/go-chi/chi/v5` (the current entry point uses `net/http` routing)
- `go-ethereum` for EVM signature verification
- `solana-go` for Solana signature verification
- `golang-jwt/jwt/v5` for OIDC ID tokens
- `golang.org/x/crypto` for bcrypt and cryptographic helpers
- `godotenv` for optional `.env` loading
- RSA RS256 signing for OIDC ID tokens

## Repository Layout

```text
cmd/main.go                 Application entry point and route wiring
config/                     Environment, database, and request logging config
database/                   PostgreSQL connection helper
migrations/                 Historical/placeholder SQL migration files
models/                     Shared domain entities
modules/
	auth/                     Wallet login and logout
	handle/                   Handle claiming and authenticated handle listing
	oidc/                     OAuth/OIDC clients, authorization, and tokens
	resolution/               Handle-to-address resolution
	session/                  OAuth session listing and revocation
	social/                   Social identity management
	user/                     Private and public profiles/dashboard
	wallet/                   Wallet linking
	wallet_list/              Authenticated wallet-list CRUD
pkg/helpers/                Signature, token, and OIDC-key helpers
pkg/middleware/             Authentication and CORS middleware
schema.sql                  Consolidated PostgreSQL bootstrap schema
new.json                    Current Postman collection
test_api.json               Older Postman collection
```

## Prerequisites

- Go `1.26.5` or a compatible Go toolchain.
- PostgreSQL 12 or newer.
- `psql` if using the provided `make migrate` target.
- A wallet capable of producing the signature format expected by the selected chain for real authentication.

## Installation and Configuration

Clone the repository, download dependencies, and create a local environment file:

```bash
go mod download
# Create .env locally, or export the variables in your shell.
```

The application calls `godotenv.Load()`. If `.env` is absent, it uses system environment variables. The tracked `.env` contains credentials and a private RSA key; treat those values as compromised, rotate them, and do not copy them into a deployment.

### Environment Variables

| Variable                   | Required        | Default                                                              | Purpose                                                                                        |
| -------------------------- | --------------- | -------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `DATABASE_URL`             | Yes in practice | `postgres://postgres:postgres@localhost:5432/nid_db?sslmode=disable` | PostgreSQL connection string                                                                   |
| `PORT`                     | No              | `8080`                                                               | HTTP listen port                                                                               |
| `FRONTEND_URL`             | No              | `http://localhost:5173`                                              | Loaded into config; current CORS middleware still uses its own localhost allowlist             |
| `NID_INTERNAL_AUTH_SECRET` | No              | `nid-secret-key`                                                     | HMAC secret for internal bearer tokens; set a strong secret                                    |
| `NID_OIDC_PRIVATE_KEY`     | No              | Temporary generated 2048-bit RSA key                                 | PEM-encoded RSA private key for OIDC ID tokens; configure a persistent key outside development |
| `NID_OIDC_ISSUER`          | No              | `http://localhost:8081`                                              | OIDC issuer used by token generation                                                           |
| `NID_OIDC_KEY_ID`          | No              | `nid-2026-01`                                                        | OIDC signing-key `kid`                                                                         |

`JWT_SECRET` appears in the supplied `.env` but is not read by the current application.

## Database Setup

Create a database and set `DATABASE_URL`:

```bash
createdb nid_db
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/nid_db?sslmode=disable'
```

The intended bootstrap command is:

```bash
make migrate
```

It runs `psql $(DATABASE_URL) -f schema.sql`. Review `schema.sql` before running it against a fresh database: the file contains legacy alterations and an `UPDATE` containing a standalone `$1` placeholder, which is not valid as a plain `psql` statement. The files under `migrations/002_wallets.sql`, `003_oauth_sessions.sql`, and `004_passkeys.sql` are empty; `001_initial_schema.sql` is an older schema that uses `handles.name`, while current application queries use the `handle` column in some paths. Do not assume the numbered files can be replayed as a working migration chain.

The current schema covers users, handles, wallets, OAuth clients/codes/tokens/sessions, OIDC signing-key metadata, social identities, and wallet-list entries. Passkey support is not implemented despite the placeholder migration filename.

## Run Locally

```bash
make run
```

The server connects to PostgreSQL during startup and listens on `:${PORT}`. With the default configuration, check it with:

```bash
curl -i http://localhost:8080/health
# HTTP/1.1 200 OK
# OK
```

The repository's workflow notes use port `8081`; set `PORT=8081` if that is your local convention. Build a deployable binary with:

```bash
make build
# bin/nid-backend
```

## API

All JSON endpoints use `Content-Type: application/json` unless noted. Protected endpoints accept `Authorization: Bearer <token>` or the `nid_token` cookie.

### Public endpoints

| Method | Path                             | Purpose                                                              |
| ------ | -------------------------------- | -------------------------------------------------------------------- |
| `GET`  | `/health`                        | Plain-text liveness response: `OK`                                   |
| `POST` | `/api/v1/auth/login`             | Verify a wallet signature and issue an internal token                |
| `POST` | `/api/v1/auth/logout`            | Clear the internal auth cookie                                       |
| `GET`  | `/api/v1/resolve?handle=&chain=` | Resolve an active handle to a wallet address                         |
| `POST` | `/api/v1/handles/claim`          | Claim a handle and create/link the wallet                            |
| `GET`  | `/api/v1/{handle}`               | Read a public profile; an initial `@` is accepted                    |
| `GET`  | `/api/v1/social/public`          | Publicly mounted route whose current handler requires authentication |

### Protected endpoints

| Method                 | Path                                       | Purpose                                                       |
| ---------------------- | ------------------------------------------ | ------------------------------------------------------------- |
| `GET`                  | `/api/v1/auth/me`                          | Current authenticated user                                    |
| `GET`                  | `/api/v1/user/profile`                     | User, active handles, and linked wallets                      |
| `GET`                  | `/api/v1/user/dashboard`                   | Aggregated handles, socials, wallet list, and active sessions |
| `GET`                  | `/api/v1/handles`                          | Authenticated user's handles                                  |
| `POST`                 | `/api/v1/wallets/link`                     | Link a wallet record                                          |
| `GET`                  | `/api/v1/sessions`                         | List OAuth sessions                                           |
| `POST`                 | `/api/v1/sessions/revoke?id=<uuid>`        | Revoke an OAuth session                                       |
| `POST/GET/PUT/DELETE`  | `/api/v1/wallet-list[/{id}]`               | Wallet-list CRUD                                              |
| `GET/POST`             | `/api/v1/social`                           | List or create social identities                              |
| `GET/PUT/PATCH/DELETE` | `/api/v1/social/{id}`                      | Read, update, toggle visibility, or delete a social identity  |
| `POST`                 | `/api/v1/oauth/register`                   | Register an OAuth client                                      |
| `GET`                  | `/api/v1/oauth/clients`                    | List the user's OAuth clients                                 |
| `DELETE`               | `/api/v1/oauth/clients/{id}`               | Delete an OAuth client                                        |
| `POST`                 | `/api/v1/oauth/clients/{id}/rotate-secret` | Rotate a confidential client secret                           |

### Claim a handle

The request DTO maps `handle` to the handle name. The service normalizes it to lowercase and expects the signed message `Claim <lowercase-handle>.nid`; the supplied `message` value is not used for this check.

```bash
curl -X POST http://localhost:8080/api/v1/handles/claim \
	-H 'Content-Type: application/json' \
	-d '{
		"handle": "alice",
		"address": "0x1234567890abcdef1234567890abcdef12345678",
		"chain": "evm",
		"signature": "<wallet-signature>",
		"message": "Claim alice.nid"
	}'
```

### Log in

```bash
curl -i -X POST http://localhost:8080/api/v1/auth/login \
	-H 'Content-Type: application/json' \
	-d '{
		"handle": "alice",
		"address": "0x1234567890abcdef1234567890abcdef12345678",
		"signature": "<wallet-signature>",
		"message": "Sign in to NID",
		"chain": "evm"
	}'
```

The response contains `token`, `user_id`, and `handle`. Use the returned token for protected requests:

```bash
curl http://localhost:8080/api/v1/user/profile \
	-H 'Authorization: Bearer <token>'
```

### Resolve an address

```bash
curl 'http://localhost:8080/api/v1/resolve?handle=alice&chain=evm'
```

Example response shape:

```json
{
	"handle": "alice",
	"chain": "evm",
	"address": "0x1234567890abcdef1234567890abcdef12345678"
}
```

## Authentication Details

Login verifies the submitted signature, confirms that the handle exists, and confirms that the wallet address is linked to the handle owner. EVM verification currently hashes the raw message; Solana signatures accept base64 or `0x`-prefixed hex and verify the exact message bytes.

Internal login tokens are deterministic HMAC values derived from the user ID. They are accepted from the bearer header or `nid_token` cookie. The cookie is HTTP-only, `SameSite=Lax`, non-secure, and lasts seven days. Internal tokens currently have no expiry, database session record, nonce, or replay protection. Wallet linking checks required fields but does not perform a new ownership-signature verification.

## OAuth 2.0 / OpenID Connect

Public OIDC endpoints are:

```text
GET  /oauth/authorize
POST /oauth/authorize/approve
GET  /oauth/client-info?client_id=<id>
POST /oauth/token
GET  /oauth/userinfo
GET  /.well-known/openid-configuration
GET  /.well-known/jwks.json
```

Register a client through the protected `POST /api/v1/oauth/register` endpoint. Clients support `public` and `confidential` types and one exact redirect URI. Authorization requires `response_type=code`, the `openid` scope, a nonce, and S256 PKCE. Authorization codes are single-use and expire after 60 seconds. Access tokens expire after one hour; ID tokens are RS256 JWTs.

<!-- SCREENSHOT: OAuth consent screen -->

The discovery handler currently advertises hardcoded `https://nid.xyz` endpoint URLs, while token signing uses `NID_OIDC_ISSUER`. Keep these values aligned for a real deployment. The controller contains a denial handler, but `/oauth/authorize/deny` is not currently registered.

## Deployment

No deployment target is included. A minimal deployment consists of:

1. Provision PostgreSQL and apply a reviewed schema.
2. Build with `make build`.
3. Inject `DATABASE_URL`, a strong `NID_INTERNAL_AUTH_SECRET`, a persistent `NID_OIDC_PRIVATE_KEY`, `NID_OIDC_ISSUER`, `NID_OIDC_KEY_ID`, `PORT`, and the appropriate frontend/CORS configuration.
4. Run `bin/nid-backend` behind a TLS-terminating reverse proxy or load balancer.
5. Expose `/health` to the platform's liveness check and keep signing keys and database credentials in a secret manager.

The application itself does not configure TLS, graceful shutdown, proxy headers, metrics, or automated database migrations.

## Troubleshooting

| Symptom                                          | Check                                                                                                                              |
| ------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------- |
| Startup fails with a database error              | Confirm PostgreSQL is running, `DATABASE_URL` is reachable, and the database schema was applied.                                   |
| `make migrate` fails                             | Inspect `schema.sql`, especially the standalone `$1` update and legacy alterations; apply a corrected bootstrap script.            |
| Requests fail with `401`                         | Send a valid bearer token or `nid_token` cookie. Login also requires a handle, matching linked address, and valid signature.       |
| Browser requests fail CORS                       | Current middleware allows only `http://localhost:5173` and `http://localhost:5174`; `FRONTEND_URL` does not expand that allowlist. |
| OIDC discovery points to the wrong host          | Review the hardcoded discovery URLs and align deployment host configuration with `NID_OIDC_ISSUER`.                                |
| OIDC tokens stop validating after restart        | Configure `NID_OIDC_PRIVATE_KEY`; without it, development startup generates a new temporary RSA key.                               |
| Client metadata is missing or says `Unknown App` | The schema contains both `name` and `client_name`; registration writes `name`, while client-info reads `client_name`.              |

## Development Notes

The current repository has no automated tests. `new.json` is the more complete Postman collection and uses the current login shape; `test_api.json`, `api_workflow.md`, `user_workflow.md`, and `arc.md` contain older or aspirational details and should not be treated as the authoritative API contract.

<!-- SCREENSHOT: Postman collection -->
