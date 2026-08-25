package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"nid-backend/config"
	"nid-backend/database"

	authCtrl "nid-backend/modules/auth/controller"
	authRepo "nid-backend/modules/auth/repository"
	authSvc "nid-backend/modules/auth/service"

	handleCtrl "nid-backend/modules/handle/controller"
	handleRepo "nid-backend/modules/handle/repository"
	handleSvc "nid-backend/modules/handle/service"

	oidcCtrl "nid-backend/modules/oidc/controller"
	oidcRepo "nid-backend/modules/oidc/repository"
	oidcSvc "nid-backend/modules/oidc/service"

	resCtrl "nid-backend/modules/resolution/controller"
	resRepo "nid-backend/modules/resolution/repository"
	resSvc "nid-backend/modules/resolution/service"

	sesCtrl "nid-backend/modules/session/controller"
	sesRepo "nid-backend/modules/session/repository"
	sesSvc "nid-backend/modules/session/service"

	userCtrl "nid-backend/modules/user/controller"
	userRepo "nid-backend/modules/user/repository"
	userSvc "nid-backend/modules/user/service"

	walletCtrl "nid-backend/modules/wallet/controller"
	walletRepo "nid-backend/modules/wallet/repository"
	walletSvc "nid-backend/modules/wallet/service"

	"nid-backend/pkg/middleware"
)

func main() {

	// ============================================================
	// Environment
	// ============================================================

	if err := godotenv.Load(); err != nil {
		log.Println(
			"No .env file found, relying on system environment variables",
		)
	}

	cfg := config.LoadConfig()

	// ============================================================
	// Database
	// ============================================================

	db, err := database.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf(
			"Failed to connect to database: %v",
			err,
		)
	}

	defer db.Close()

	// ============================================================
	// Repositories
	// ============================================================

	authRepoInstance :=
		authRepo.NewAuthRepository(db)

	handleRepoInstance :=
		handleRepo.NewHandleRepository(db)

	walletRepoInstance :=
		walletRepo.NewWalletRepository(db)

	resolutionRepoInstance :=
		resRepo.NewResolutionRepository(db)

	sessionRepoInstance :=
		sesRepo.NewSessionRepository(db)

	userRepoInstance :=
		userRepo.NewUserRepository(db)

	oidcRepoInstance :=
		oidcRepo.NewOIDCRepository(db)

	// ============================================================
	// OIDC Signing Key
	// ============================================================

	oidcPrivateKey, err := loadOIDCPrivateKey()
	if err != nil {
		log.Fatalf(
			"Failed to load OIDC private key: %v",
			err,
		)
	}

	// ============================================================
	// Services
	// ============================================================

	authService :=
		authSvc.NewAuthService(
			authRepoInstance,
		)

	handleService :=
		handleSvc.NewHandleService(
			handleRepoInstance,
		)

	walletService :=
		walletSvc.NewWalletService(
			walletRepoInstance,
		)

	resolutionService :=
		resSvc.NewResolutionService(
			resolutionRepoInstance,
		)

	sessionService :=
		sesSvc.NewSessionService(
			sessionRepoInstance,
		)

	userService :=
		userSvc.NewUserService(
			userRepoInstance,
		)

	// ============================================================
	// OIDC Configuration
	// ============================================================

	oidcIssuer := getEnv(
		"NID_OIDC_ISSUER",
		"http://localhost:8081",
	)

	oidcKeyID := getEnv(
		"NID_OIDC_KEY_ID",
		"nid-2026-01",
	)

	oidcService :=
		oidcSvc.NewOIDCService(
			oidcRepoInstance,
			oidcPrivateKey,
			oidcIssuer,
			oidcKeyID,
		)

	// ============================================================
	// Controllers
	// ============================================================

	authController :=
		authCtrl.NewAuthController(
			authService,
		)

	handleController :=
		handleCtrl.NewHandleController(
			handleService,
		)

	walletController :=
		walletCtrl.NewWalletController(
			walletService,
		)

	resolutionController :=
		resCtrl.NewResolutionController(
			resolutionService,
		)

	sessionController :=
		sesCtrl.NewSessionController(
			sessionService,
		)

	userController :=
		userCtrl.NewUserController(
			userService,
		)

	oidcController :=
		oidcCtrl.NewOIDCController(
			oidcService,
		)

	// ============================================================
	// Main Router
	// ============================================================

	mux := http.NewServeMux()

	// ============================================================
	// 1. Health
	// ============================================================

	mux.HandleFunc(
		"/health",
		func(w http.ResponseWriter, r *http.Request) {

			if r.Method != http.MethodGet {
				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)
				return
			}

			w.Header().Set(
				"Content-Type",
				"text/plain; charset=utf-8",
			)

			w.WriteHeader(http.StatusOK)

			_, _ = w.Write(
				[]byte("OK"),
			)
		},
	)

	// ============================================================
	// 2. Core Public API
	// ============================================================

	mux.HandleFunc(
		"/api/v1/auth/login",
		authController.LoginHandler,
	)

	mux.HandleFunc(
		"/api/v1/resolve",
		resolutionController.ResolveHandler,
	)

	mux.HandleFunc(
		"/api/v1/handles/claim",
		handleController.ClaimHandler,
	)

	// ============================================================
	// 3. OAuth 2.0 / OpenID Connect
	// ============================================================

	// ------------------------------------------------------------
	// Client Registration
	// ------------------------------------------------------------

	mux.HandleFunc(
		"/oauth/register",
		oidcController.RegisterClientHandler,
	)

	// ------------------------------------------------------------
	// Authorization Endpoint
	//
	// GET /oauth/authorize
	//
	// Client sends:
	//
	// client_id
	// redirect_uri
	// response_type
	// scope
	// state
	// nonce
	// code_challenge
	// code_challenge_method
	// ------------------------------------------------------------

	mux.HandleFunc(
		"/oauth/authorize",
		oidcController.AuthorizeHandler,
	)

	// ------------------------------------------------------------
	// Authorization Approval
	//
	// POST /oauth/authorize/approve
	//
	// IMPORTANT:
	// Your NID authorization frontend calls this endpoint.
	// ------------------------------------------------------------

	mux.HandleFunc(
		"/oauth/authorize/approve",
		oidcController.ApproveAuthorizationHandler,
	)

	// ------------------------------------------------------------
	// Token Endpoint
	//
	// POST /oauth/token
	// ------------------------------------------------------------

	mux.HandleFunc(
		"/oauth/token",
		oidcController.TokenHandler,
	)

	// ------------------------------------------------------------
	// UserInfo Endpoint
	//
	// GET /oauth/userinfo
	// ------------------------------------------------------------

	mux.HandleFunc(
		"/oauth/userinfo",
		oidcController.UserInfoHandler,
	)

	// ============================================================
	// 4. OpenID Connect Discovery
	// ============================================================

	mux.HandleFunc(
		"/.well-known/openid-configuration",
		oidcController.DiscoveryHandler,
	)

	// ============================================================
	// 5. JWKS
	// ============================================================

	mux.HandleFunc(
		"/.well-known/jwks.json",
		oidcController.JWKSHandler,
	)

	// ============================================================
	// 6. Protected Routes
	// ============================================================

	protectedMux := http.NewServeMux()

	// ------------------------------------------------------------
	// Link wallet
	// ------------------------------------------------------------

	protectedMux.HandleFunc(
		"/api/v1/wallets/link",
		walletController.LinkWalletHandler,
	)

	// ------------------------------------------------------------
	// Revoke session
	// ------------------------------------------------------------

	protectedMux.HandleFunc(
		"/api/v1/sessions/revoke",
		sessionController.RevokeHandler,
	)

	// ------------------------------------------------------------
	// User profile
	// ------------------------------------------------------------

	protectedMux.HandleFunc(
		"/api/v1/user/profile",
		userController.GetProfileHandler,
	)

	// ============================================================
	// 7. Authentication Middleware
	// ============================================================

	mux.Handle(
		"/api/v1/wallets/",
		middleware.AuthMiddleware(
			protectedMux,
		),
	)

	mux.Handle(
		"/api/v1/sessions/",
		middleware.AuthMiddleware(
			protectedMux,
		),
	)

	mux.Handle(
		"/api/v1/user/",
		middleware.AuthMiddleware(
			protectedMux,
		),
	)

	// ============================================================
	// 8. Global Middleware
	// ============================================================

	handler := middleware.CORSMiddleware(
		config.RequestLogger(mux),
	)

	// ============================================================
	// 9. HTTP Server
	// ============================================================

	port := cfg.Port

	if strings.TrimSpace(port) == "" {
		port = "8081"
	}

	log.Printf(
		"NID backend starting on port %s",
		port,
	)

	log.Printf(
		"OIDC issuer: %s",
		oidcIssuer,
	)

	log.Printf(
		"OIDC authorization endpoint: %s/oauth/authorize",
		oidcIssuer,
	)

	log.Printf(
		"OIDC approval endpoint: %s/oauth/authorize/approve",
		oidcIssuer,
	)

	log.Printf(
		"OIDC token endpoint: %s/oauth/token",
		oidcIssuer,
	)

	log.Printf(
		"OIDC userinfo endpoint: %s/oauth/userinfo",
		oidcIssuer,
	)

	log.Printf(
		"OIDC discovery endpoint: %s/.well-known/openid-configuration",
		oidcIssuer,
	)

	log.Printf(
		"OIDC JWKS endpoint: %s/.well-known/jwks.json",
		oidcIssuer,
	)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,

		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Fatalf(
				"Server failed: %v",
				err,
			)
		}
	}
}

// ============================================================
// Environment Helper
// ============================================================

func getEnv(
	key string,
	fallback string,
) string {

	value := strings.TrimSpace(
		os.Getenv(key),
	)

	if value == "" {
		return fallback
	}

	return value
}

// ============================================================
// Load OIDC Private Key
// ============================================================
//
// Expected environment variable:
//
// NID_OIDC_PRIVATE_KEY
//
// It can contain:
//
// -----BEGIN RSA PRIVATE KEY-----
// ...
// -----END RSA PRIVATE KEY-----
//
// or:
//
// -----BEGIN PRIVATE KEY-----
// ...
// -----END PRIVATE KEY-----
//
// ============================================================

func loadOIDCPrivateKey() (*rsa.PrivateKey, error) {

	privateKeyPEM := strings.TrimSpace(
		os.Getenv("NID_OIDC_PRIVATE_KEY"),
	)

	// ------------------------------------------------------------
	// Development fallback
	// ------------------------------------------------------------

	if privateKeyPEM == "" {

		log.Println(
			"WARNING: NID_OIDC_PRIVATE_KEY is not configured",
		)

		log.Println(
			"Generating temporary RSA key for development",
		)

		return rsa.GenerateKey(
			rand.Reader,
			2048,
		)
	}

	// ------------------------------------------------------------
	// Decode PEM
	// ------------------------------------------------------------

	block, _ := pem.Decode(
		[]byte(privateKeyPEM),
	)

	if block == nil {
		return nil, os.ErrInvalid
	}

	// ------------------------------------------------------------
	// PKCS#1
	// ------------------------------------------------------------

	if key, err := x509.ParsePKCS1PrivateKey(
		block.Bytes,
	); err == nil {

		return key, nil
	}

	// ------------------------------------------------------------
	// PKCS#8
	// ------------------------------------------------------------

	key, err := x509.ParsePKCS8PrivateKey(
		block.Bytes,
	)

	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)

	if !ok {
		return nil, os.ErrInvalid
	}

	return rsaKey, nil
}
