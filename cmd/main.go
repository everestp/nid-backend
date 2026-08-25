package main

import (
	"errors"
	"log"
	"net/http"
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

	socialCtrl "nid-backend/modules/social/controller"
	socialRepo "nid-backend/modules/social/repository"
	socialSvc "nid-backend/modules/social/service"

	userCtrl "nid-backend/modules/user/controller"
	userRepo "nid-backend/modules/user/repository"
	userSvc "nid-backend/modules/user/service"

	walletCtrl "nid-backend/modules/wallet/controller"
	walletRepo "nid-backend/modules/wallet/repository"
	walletSvc "nid-backend/modules/wallet/service"

	"nid-backend/pkg/helpers"
	"nid-backend/pkg/middleware"
)

func main() {

	// ============================================================
	// Environment
	// ============================================================

	if err := godotenv.Load(); err != nil {
		log.Println(
			"No .env file found, using system environment variables",
		)
	}

	cfg := config.LoadConfig()

	// ============================================================
	// Database
	// ============================================================

	db, err := database.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf(
			"failed to connect to database: %v",
			err,
		)
	}

	defer db.Close()

	log.Println("Database connected successfully")

	// ============================================================
	// Repositories
	// ============================================================

	authRepository :=
		authRepo.NewAuthRepository(db)

	handleRepository :=
		handleRepo.NewHandleRepository(db)

	walletRepository :=
		walletRepo.NewWalletRepository(db)

	resolutionRepository :=
		resRepo.NewResolutionRepository(db)

	sessionRepository :=
		sesRepo.NewSessionRepository(db)

	userRepository :=
		userRepo.NewUserRepository(db)

	oidcRepository :=
		oidcRepo.NewOIDCRepository(db)

	socialRepository :=
		socialRepo.NewSocialRepository(db)

	// ============================================================
	// OIDC Signing Key
	// ============================================================

	oidcPrivateKey, err :=
		helpers.LoadOIDCPrivateKey()

	if err != nil {
		log.Fatalf(
			"failed to load OIDC private key: %v",
			err,
		)
	}

	// ============================================================
	// Services
	// ============================================================

	authService :=
		authSvc.NewAuthService(
			authRepository,
		)

	handleService :=
		handleSvc.NewHandleService(
			handleRepository,
		)

	walletService :=
		walletSvc.NewWalletService(
			walletRepository,
		)

	resolutionService :=
		resSvc.NewResolutionService(
			resolutionRepository,
		)

	sessionService :=
		sesSvc.NewSessionService(
			sessionRepository,
		)

	userService :=
		userSvc.NewUserService(
			userRepository,
		)

	socialService :=
		socialSvc.NewSocialService(
			socialRepository,
		)

	// ============================================================
	// OIDC Configuration
	// ============================================================

	oidcIssuer := config.GetEnv(
		"NID_OIDC_ISSUER",
		"http://localhost:8081",
	)

	oidcKeyID := config.GetEnv(
		"NID_OIDC_KEY_ID",
		"nid-2026-01",
	)

	oidcService :=
		oidcSvc.NewOIDCService(
			oidcRepository,
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

	socialController :=
		socialCtrl.NewSocialController(
			socialService,
		)

	// ============================================================
	// MAIN ROUTER
	// ============================================================

	mux := http.NewServeMux()

	// ============================================================
	// HEALTH
	// ============================================================

	mux.HandleFunc(
		"GET /health",
		healthHandler,
	)

	// ============================================================
	// PUBLIC API
	// ============================================================

	// ============================================================
	// In-House Authentication
	// ============================================================

	mux.HandleFunc(
		"POST /api/v1/auth/login",
		authController.LoginHandler,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/logout",
		authController.LogoutHandler,
	)

	// ============================================================
	// Handle Resolution
	// ============================================================

	mux.HandleFunc(
		"GET /api/v1/resolve",
		resolutionController.ResolveHandler,
	)

	// ============================================================
	// Claim Handle
	// ============================================================

	mux.HandleFunc(
		"POST /api/v1/handles/claim",
		handleController.ClaimHandler,
	)

	// ============================================================
	// PUBLIC SOCIAL PROFILE
	// ============================================================

	mux.HandleFunc(
		"GET /api/v1/social/public",
		socialController.PublicListHandler,
	)

	// ============================================================
	// OAUTH 2.0 / OPENID CONNECT
	// ============================================================

	// ============================================================
	// Register OAuth Client
	// ============================================================

	mux.HandleFunc(
		"POST /oauth/register",
		oidcController.RegisterClientHandler,
	)

	// ============================================================
	// Authorization Endpoint
	// ============================================================

	mux.HandleFunc(
		"GET /oauth/authorize",
		oidcController.AuthorizeHandler,
	)

	// ============================================================
	// Authorization Approval
	// ============================================================

	mux.HandleFunc(
		"POST /oauth/authorize/approve",
		oidcController.ApproveAuthorizationHandler,
	)

	// ============================================================
	// Client Information
	// ============================================================

	mux.HandleFunc(
		"GET /oauth/client-info",
		oidcController.GetClientInfoHandler,
	)

	// ============================================================
	// Token Endpoint
	// ============================================================

	mux.HandleFunc(
		"POST /oauth/token",
		oidcController.TokenHandler,
	)

	// ============================================================
	// UserInfo Endpoint
	// ============================================================

	mux.HandleFunc(
		"GET /oauth/userinfo",
		oidcController.UserInfoHandler,
	)

	// ============================================================
	// OIDC Discovery
	// ============================================================

	mux.HandleFunc(
		"GET /.well-known/openid-configuration",
		oidcController.DiscoveryHandler,
	)

	// ============================================================
	// OIDC JWKS
	// ============================================================

	mux.HandleFunc(
		"GET /.well-known/jwks.json",
		oidcController.JWKSHandler,
	)

	// ============================================================
	// PROTECTED API
	// ============================================================

	protectedMux := http.NewServeMux()

	// ============================================================
	// Wallet
	// ============================================================

	protectedMux.HandleFunc(
		"POST /api/v1/wallets/link",
		walletController.LinkWalletHandler,
	)

	// ============================================================
	// Sessions
	// ============================================================

	// List all sessions for current user
	//
	// GET /api/v1/sessions
	//
	protectedMux.HandleFunc(
		"GET /api/v1/sessions",
		sessionController.ListHandler,
	)

	// Revoke one session
	//
	// POST /api/v1/sessions/revoke?id=<session-id>
	//
	protectedMux.HandleFunc(
		"POST /api/v1/sessions/revoke",
		sessionController.RevokeHandler,
	)

	// ============================================================
	// User
	// ============================================================

	protectedMux.HandleFunc(
		"GET /api/v1/user/profile",
		userController.GetProfileHandler,
	)

	// ============================================================
	// Social
	// ============================================================

	// Get current user's social identities
	protectedMux.HandleFunc(
		"GET /api/v1/social",
		socialController.ListHandler,
	)

	// Get one social identity
	protectedMux.HandleFunc(
		"GET /api/v1/social/{id}",
		socialController.GetHandler,
	)

	// Add social identity
	protectedMux.HandleFunc(
		"POST /api/v1/social",
		socialController.CreateHandler,
	)

	// Update social identity
	protectedMux.HandleFunc(
		"PUT /api/v1/social/{id}",
		socialController.UpdateHandler,
	)

	// Toggle visibility
	protectedMux.HandleFunc(
		"PATCH /api/v1/social/{id}/visibility",
		socialController.ToggleVisibilityHandler,
	)

	// Delete social identity
	protectedMux.HandleFunc(
		"DELETE /api/v1/social/{id}",
		socialController.DeleteHandler,
	)

	// ============================================================
	// Protected Middleware
	// ============================================================

	protectedHandler :=
		middleware.AuthMiddleware(
			protectedMux,
		)

	// ============================================================
	// Mount Protected Wallet API
	// ============================================================

	mux.Handle(
		"/api/v1/wallets/",
		protectedHandler,
	)

	// ============================================================
	// Mount Protected Session API
	// ============================================================

	mux.Handle(
		"/api/v1/sessions",
		protectedHandler,
	)

	mux.Handle(
		"/api/v1/sessions/",
		protectedHandler,
	)

	// ============================================================
	// Mount Protected User API
	// ============================================================

	mux.Handle(
		"/api/v1/user/",
		protectedHandler,
	)

	// ============================================================
	// Mount Protected Social API
	// ============================================================

	mux.Handle(
		"/api/v1/social",
		protectedHandler,
	)

	mux.Handle(
		"/api/v1/social/",
		protectedHandler,
	)

	// ============================================================
	// Global Middleware
	// ============================================================

	handler :=
		middleware.CORSMiddleware(
			config.RequestLogger(
				mux,
			),
		)

	// ============================================================
	// HTTP SERVER
	// ============================================================

	port := strings.TrimSpace(cfg.Port)

	if port == "" {
		port = "8081"
	}

	server := &http.Server{
		Addr: ":" + port,

		Handler: handler,

		ReadHeaderTimeout: 10 * time.Second,

		ReadTimeout: 15 * time.Second,

		WriteTimeout: 15 * time.Second,

		IdleTimeout: 60 * time.Second,
	}

	// ============================================================
	// STARTUP LOGS
	// ============================================================

	log.Println("==============================================")
	log.Println("NID Backend")
	log.Println("==============================================")

	log.Printf(
		"HTTP server: http://localhost:%s",
		port,
	)

	log.Printf(
		"Health: http://localhost:%s/health",
		port,
	)

	log.Printf(
		"OIDC issuer: %s",
		oidcIssuer,
	)

	log.Printf(
		"OIDC authorize: %s/oauth/authorize",
		oidcIssuer,
	)

	log.Printf(
		"OIDC token: %s/oauth/token",
		oidcIssuer,
	)

	log.Printf(
		"OIDC userinfo: %s/oauth/userinfo",
		oidcIssuer,
	)

	log.Printf(
		"OIDC discovery: %s/.well-known/openid-configuration",
		oidcIssuer,
	)

	log.Printf(
		"OIDC JWKS: %s/.well-known/jwks.json",
		oidcIssuer,
	)

	log.Println("==============================================")

	// ============================================================
	// START SERVER
	// ============================================================

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(
			err,
			http.ErrServerClosed,
		) {
			log.Fatalf(
				"server failed: %v",
				err,
			)
		}
	}
}

// ============================================================
// HEALTH HANDLER
// ============================================================

func healthHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"text/plain; charset=utf-8",
	)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(
		[]byte("OK"),
	)
}
