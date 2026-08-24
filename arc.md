nid-backend/
├── cmd/
│   └── main.go                  # Application entry point; initializes server, router, and DB connections
├── config/
│   ├── database.go              # Database configuration (PostgreSQL connection pool setup)
│   └── logger.go                # Centralized logging setup for tracking requests and errors
├── database/
│   └── db.go                    # Global database helper functions and instance management
├── migrations/
│   ├── 001_initial_schema.sql   # Database migration: Users and base tables
│   ├── 002_wallets.sql          # Database migration: Multi-chain wallet bindings
│   ├── 003_oauth_sessions.sql   # Database migration: OAuth/OIDC tokens and sessions
│   └── 004_passkeys.sql         # Database migration: FIDO2 passkeys for passwordless auth
├── models/
│   └── entity.go                # Shared domain models and core database entities
├── modules/
│   ├── auth/                    # Authentication module (nonce generation, wallet signature verification, JWT issue)
│   │   ├── controller/          # Handles HTTP requests for login, authorize, and token exchange
│   │   ├── dto/                 # Data Transfer Objects for auth requests and responses
│   │   ├── repository/          # Database queries related to authentication state and challenges
│   │   └── service/             # Business logic for cryptographic validation and token generation
│   ├── handle/                  # Namespace module (manages .nid identity handles)
│   │   ├── controller/          # Handles endpoints for claiming and looking up .nid handles
│   │   ├── dto/                 # Data Transfer Objects for handle registration and search
│   │   ├── repository/          # Database queries for handle availability and ownership
│   │   └── service/             # Business logic for namespace creation and validation rules
│   ├── wallet/                  # Wallet binding module (links EVM & Solana addresses to user accounts)
│   │   ├── controller/          # Handles linking and unlinking wallet endpoints
│   │   ├── dto/                 # Data Transfer Objects for wallet bindings
│   │   ├── repository/          # Database queries for storing and fetching user wallets
│   │   └── service/             # Business logic for multi-chain address management
│   ├── resolution/              # Address resolution module (handles forward & reverse .nid lookups for crypto transfers)
│   │   ├── controller/          # Handles API endpoints for resolving .nid handles to chain addresses
│   │   ├── dto/                 # Data Transfer Objects for resolution queries
│   │   ├── repository/          # Database queries to fetch addresses mapped to specific chains
│   │   └── service/             # Business logic for multi-chain address resolution mapping
│   ├── session/                 # Active sessions & connected apps module (manages third-party app access)
│   │   ├── controller/          # Handles listing active app sessions and token revocation
│   │   ├── dto/                 # Data Transfer Objects for session management
│   │   ├── repository/          # Database queries for tracking active OAuth/OIDC grants
│   │   └── service/             # Business logic for session lifecycles and permission revocation
│   └── user/                    # User profile module (manages account settings and preferences)
│       ├── controller/          # Handles user profile retrieval and account updates
│       ├── dto/                 # Data Transfer Objects for user profiles
│       ├── repository/          # Database queries for user account data
│       └── service/             # Business logic for user account management
├── pkg/
│   ├── helpers/
│   │   └── auth.go              # Utility functions for generating and verifying JWTs and signatures
│   ├── middleware/
│   │   ├── auth_middleware.go   # HTTP middleware to validate bearer tokens on protected routes
│   │   └── cors.go              # HTTP middleware for Cross-Origin Resource Sharing settings
│   └── utils/
│       └── response.go          # Standardized JSON response wrapper for success and error payloads
├── go.mod                       # Go module dependencies definition file
├── go.sum                       # Checksums for Go module dependencies
├── Makefile                     # Build, test, run, and migration task runner commands
└── schema.sql                   # Fully consolidated database schema file for fresh setups
