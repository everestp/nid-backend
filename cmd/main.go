package main

import (
	"crypto/rsa"
	"log"
	"net/http"

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
		log.Println("No .env file found, relying on system environment variables")
	}

	cfg := config.LoadConfig()

	// ============================================================
	// Database
	// ============================================================

	db, err := database.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()

	// ============================================================
	// Repositories
	// ============================================================

	authRepoInstance := authRepo.NewAuthRepository(db)

	handleRepoInstance := handleRepo.NewHandleRepository(db)

	walletRepoInstance := walletRepo.NewWalletRepository(db)

	resolutionRepoInstance := resRepo.NewResolutionRepository(db)

	sessionRepoInstance := sesRepo.NewSessionRepository(db)

	userRepoInstance := userRepo.NewUserRepository(db)

	oidcRepoInstance := oidcRepo.NewOIDCRepository(db)
var oidcPrivateKey *rsa.PrivateKey
	// ============================================================
	// Services
	// ============================================================

	authService := authSvc.NewAuthService(authRepoInstance)

	handleService := handleSvc.NewHandleService(handleRepoInstance)

	walletService := walletSvc.NewWalletService(walletRepoInstance)

	resolutionService := resSvc.NewResolutionService(
		resolutionRepoInstance,
	)

	sessionService := sesSvc.NewSessionService(
		sessionRepoInstance,
	)

	userService := userSvc.NewUserService(
		userRepoInstance,
	)

oidcService := oidcSvc.NewOIDCService(
	oidcRepoInstance,
	oidcPrivateKey,
	"https://nid.xyz",
	"nid-2026-01",
)

	// ============================================================
	// Controllers
	// ============================================================

	authController := authCtrl.NewAuthController(
		authService,
	)

	handleController := handleCtrl.NewHandleController(
		handleService,
	)

	walletController := walletCtrl.NewWalletController(
		walletService,
	)

	resolutionController := resCtrl.NewResolutionController(
		resolutionService,
	)

	sessionController := sesCtrl.NewSessionController(
		sessionService,
	)

	userController := userCtrl.NewUserController(
		userService,
	)

	oidcController := oidcCtrl.NewOIDCController(
		oidcService,
	)

	// ============================================================
	// Main Router
	// ============================================================

	mux := http.NewServeMux()

	// ============================================================
	// 1. Health
	// ============================================================

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
			return
		}

		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("OK"))
	})

	// ============================================================
	// 2. Core Public Routes
	// ============================================================

	// Wallet login / NID login
	mux.HandleFunc(
		"/api/v1/auth/login",
		authController.LoginHandler,
	)

	// Resolve .nid handle
	mux.HandleFunc(
		"/api/v1/resolve",
		resolutionController.ResolveHandler,
	)

	// Claim .nid handle
	mux.HandleFunc(
		"/api/v1/handles/claim",
		handleController.ClaimHandler,
	)

	// ============================================================
	// 3. OAuth 2.0 / OpenID Connect
	// ============================================================

	// ------------------------------------------------------------
	// Dynamic Client Registration
	//
	// Third-party applications can register themselves here.
	//
	// POST /oauth/register
	//
	// Returns:
	// {
	//   "client_id": "...",
	//   "client_secret": "...",
	//   ...
	// }
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
	// Example:
	//
	// /oauth/authorize?
	//   client_id=...
	//   &redirect_uri=https://app.com/callback
	//   &response_type=code
	//   &scope=openid
	//   &state=xyz
	// ------------------------------------------------------------

	mux.HandleFunc(
		"/oauth/authorize",
		oidcController.AuthorizeHandler,
	)

	// ------------------------------------------------------------
	// Token Endpoint
	//
	// POST /oauth/token
	//
	// Exchanges authorization code for:
	// - access_token
	// - id_token
	// ------------------------------------------------------------

	mux.HandleFunc(
		"/oauth/token",
		oidcController.TokenHandler,
	)

	// ------------------------------------------------------------
	// UserInfo Endpoint
	//
	// GET /oauth/userinfo
	//
	// Used by OIDC clients to obtain user claims.
	// ------------------------------------------------------------

	mux.HandleFunc(
		"/oauth/userinfo",
		oidcController.UserInfoHandler,
	)

	// ------------------------------------------------------------
	// OpenID Connect Discovery
	//
	// GET /.well-known/openid-configuration
	// ------------------------------------------------------------

	mux.HandleFunc(
		"/.well-known/openid-configuration",
		oidcController.DiscoveryHandler,
	)

	// ============================================================
	// 4. Protected Routes
	// ============================================================

	protectedMux := http.NewServeMux()

	// Link wallet
	protectedMux.HandleFunc(
		"/api/v1/wallets/link",
		walletController.LinkWalletHandler,
	)

	// Revoke session
	protectedMux.HandleFunc(
		"/api/v1/sessions/revoke",
		sessionController.RevokeHandler,
	)

	// User profile
	protectedMux.HandleFunc(
		"/api/v1/user/profile",
		userController.GetProfileHandler,
	)

	// ============================================================
	// 5. Authentication Middleware
	// ============================================================

	mux.Handle(
		"/api/v1/wallets/",
		middleware.AuthMiddleware(protectedMux),
	)

	mux.Handle(
		"/api/v1/sessions/",
		middleware.AuthMiddleware(protectedMux),
	)

	mux.Handle(
		"/api/v1/user/",
		middleware.AuthMiddleware(protectedMux),
	)

	// ============================================================
	// 6. Global Middleware
	// ============================================================

	handler := middleware.CORSMiddleware(
		config.RequestLogger(mux),
	)

	// ============================================================
	// 7. Start Server
	// ============================================================

	log.Printf(
		"Server starting on port %s...",
		cfg.Port,
	)

	if err := http.ListenAndServe(
		":"+cfg.Port,
		handler,
	); err != nil {

		log.Fatalf(
			"Server failed: %v",
			err,
		)
	}
}
