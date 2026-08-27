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

	walletListCtrl "nid-backend/modules/wallet_list/controller"
	walletListRepo "nid-backend/modules/wallet_list/repository"
	walletListSvc "nid-backend/modules/wallet_list/service"

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
		log.Println("No .env file found, using system environment variables")
	}

	cfg := config.LoadConfig()

	// ============================================================
	// Database
	// ============================================================

	db, err := database.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	// ============================================================
	// Repositories
	// ============================================================

	authRepository := authRepo.NewAuthRepository(db)
	handleRepository := handleRepo.NewHandleRepository(db)
	walletRepository := walletRepo.NewWalletRepository(db)
	resolutionRepository := resRepo.NewResolutionRepository(db)
	sessionRepository := sesRepo.NewSessionRepository(db)
	userRepository := userRepo.NewUserRepository(db)
	oidcRepository := oidcRepo.NewOIDCRepository(db)
	socialRepository := socialRepo.NewSocialRepository(db)
	walletListRepository := walletListRepo.NewwalletListRepository(db)

	// ============================================================
	// OIDC Signing Key
	// ============================================================

	oidcPrivateKey, err := helpers.LoadOIDCPrivateKey()
	if err != nil {
		log.Fatalf("failed to load OIDC private key: %v", err)
	}

	// ============================================================
	// Services
	// ============================================================

	authService := authSvc.NewAuthService(authRepository)
	handleService := handleSvc.NewHandleService(handleRepository)
	walletService := walletSvc.NewWalletService(walletRepository)
	resolutionService := resSvc.NewResolutionService(resolutionRepository)
	sessionService := sesSvc.NewSessionService(sessionRepository)
	userService := userSvc.NewUserService(userRepository)
	socialService := socialSvc.NewSocialService(socialRepository)
	walletListService := walletListSvc.NewwalletListService(walletListRepository)

	// ============================================================
	// OIDC Configuration
	// ============================================================

	oidcIssuer := config.GetEnv(
		"NID_OIDC_ISSUER",
		"https://api.nid.xyz",
	)

	oidcKeyID := config.GetEnv(
		"NID_OIDC_KEY_ID",
		"nid-2026-01",
	)

	oidcService := oidcSvc.NewOIDCService(
		oidcRepository,
		oidcPrivateKey,
		oidcIssuer,
		oidcKeyID,
	)

	// ============================================================
	// Controllers
	// ============================================================

	authController := authCtrl.NewAuthController(authService)
	handleController := handleCtrl.NewHandleController(handleService)
	walletController := walletCtrl.NewWalletController(walletService)
	resolutionController := resCtrl.NewResolutionController(resolutionService)
	sessionController := sesCtrl.NewSessionController(sessionService)
	userController := userCtrl.NewUserController(userService)
	oidcController := oidcCtrl.NewOIDCController(oidcService)
	socialController := socialCtrl.NewSocialController(socialService)
	walletListController := walletListCtrl.NewwalletListtController(walletListService)

	// ============================================================
	// MAIN ROUTER
	// ============================================================

	mux := http.NewServeMux()

	// ============================================================
	// HEALTH
	// ============================================================

	mux.HandleFunc("GET /health", healthHandler)

	// ============================================================
	// PUBLIC API
	// ============================================================

	// Auth
	mux.HandleFunc(
		"POST /api/v1/auth/login",
		authController.LoginHandler,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/logout",
		authController.LogoutHandler,
	)

	// Handle Resolution
	mux.HandleFunc(
		"GET /api/v1/resolve",
		resolutionController.ResolveHandler,
	)

	// Claim Handle
	mux.HandleFunc(
		"POST /api/v1/handles/claim",
		handleController.ClaimHandler,
	)

	// Public Social Profile
	mux.HandleFunc(
		"GET /api/v1/social/public",
		socialController.PublicListHandler,
	)

	// Public User Profile
	//
	// GET /api/v1/alice
	// GET /api/v1/@alice
	//
	mux.HandleFunc(
		"GET /api/v1/{handle}",
		userController.GetPublicProfileByHandleHandler,
	)

	// ============================================================
	// PUBLIC OAUTH / OPENID CONNECT
	// ============================================================

	mux.HandleFunc(
		"GET /oauth/authorize",
		oidcController.AuthorizeHandler,
	)

	mux.HandleFunc(
		"POST /oauth/authorize/approve",
		oidcController.ApproveAuthorizationHandler,
	)

	mux.HandleFunc(
		"GET /oauth/client-info",
		oidcController.GetClientInfoHandler,
	)

	mux.HandleFunc(
		"POST /oauth/token",
		oidcController.TokenHandler,
	)

	mux.HandleFunc(
		"GET /oauth/userinfo",
		oidcController.UserInfoHandler,
	)

	mux.HandleFunc(
		"GET /.well-known/openid-configuration",
		oidcController.DiscoveryHandler,
	)

	mux.HandleFunc(
		"GET /.well-known/jwks.json",
		oidcController.JWKSHandler,
	)

	// ============================================================
	// PROTECTED ROUTE HELPER
	// ============================================================
	//
	// Every protected endpoint goes through AuthMiddleware.
	//
	// This avoids mounting multiple overlapping prefixes on the
	// main ServeMux.
	//

	protected := func(handler http.HandlerFunc) http.HandlerFunc {
		return middleware.AuthMiddleware(
			http.HandlerFunc(handler),
		).ServeHTTP
	}

	// ============================================================
	// PROTECTED - OAUTH CLIENTS
	// ============================================================

	mux.HandleFunc(
		"POST /api/v1/oauth/register",
		protected(oidcController.RegisterClientHandler),
	)

	mux.HandleFunc(
		"GET /api/v1/oauth/clients",
		protected(oidcController.ListAllByUserHandler),
	)

	mux.HandleFunc(
		"DELETE /api/v1/oauth/clients/{id}",
		protected(oidcController.DeleteClientHandler),
	)

	mux.HandleFunc(
		"POST /api/v1/oauth/clients/{id}/rotate-secret",
		protected(oidcController.RotateClientSecretHandler),
	)
// Get All Handles By Authenticated User
mux.HandleFunc(
	"GET /api/v1/handles",
protected(handleController.GetAllByUserIDHandler),
)
	// ============================================================
	// PROTECTED - WALLET LIST
	// ============================================================

	mux.HandleFunc(
		"POST /api/v1/wallet-list",
		protected(walletListController.Create),
	)

	mux.HandleFunc(
		"GET /api/v1/wallet-list",
		protected(walletListController.GetAll),
	)

	mux.HandleFunc(
		"GET /api/v1/wallet-list/{id}",
		protected(walletListController.GetByID),
	)

	mux.HandleFunc(
		"PUT /api/v1/wallet-list/{id}",
		protected(walletListController.Update),
	)

	mux.HandleFunc(
		"DELETE /api/v1/wallet-list/{id}",
		protected(walletListController.Delete),
	)

	// ============================================================
	// PROTECTED - WALLET
	// ============================================================

	mux.HandleFunc(
		"POST /api/v1/wallets/link",
		protected(walletController.LinkWalletHandler),
	)

	// ============================================================
	// PROTECTED - SESSIONS
	// ============================================================

	mux.HandleFunc(
		"GET /api/v1/sessions",
		protected(sessionController.ListHandler),
	)

	mux.HandleFunc(
		"POST /api/v1/sessions/revoke",
		protected(sessionController.RevokeHandler),
	)

	// ============================================================
	// PROTECTED - USER
	// ============================================================

	mux.HandleFunc(
		"GET /api/v1/user/profile",
		protected(userController.GetProfileHandler),
	)
	mux.HandleFunc(
    "GET /api/v1/user/dashboard",
    protected(userController.GetDashboardHandler),
)
mux.HandleFunc(
	"GET /api/v1/auth/me",
	protected(userController.GetCurrentLoggedInUserHandler),
)
	// ============================================================
	// PROTECTED - SOCIAL
	// ============================================================

	mux.HandleFunc(
		"GET /api/v1/social",
		protected(socialController.ListHandler),
	)

	mux.HandleFunc(
		"GET /api/v1/social/{id}",
		protected(socialController.GetHandler),
	)

	mux.HandleFunc(
		"POST /api/v1/social",
		protected(socialController.CreateHandler),
	)

	mux.HandleFunc(
		"PUT /api/v1/social/{id}",
		protected(socialController.UpdateHandler),
	)

	mux.HandleFunc(
		"PATCH /api/v1/social/{id}/visibility",
		protected(socialController.ToggleVisibilityHandler),
	)

	mux.HandleFunc(
		"DELETE /api/v1/social/{id}",
		protected(socialController.DeleteHandler),
	)

	// ============================================================
	// GLOBAL MIDDLEWARE
	// ============================================================

	handler := middleware.CORSMiddleware(
		config.RequestLogger(mux),
	)

	// ============================================================
	// HTTP SERVER
	// ============================================================

	port := strings.TrimSpace(cfg.Port)

	if port == "" {
		port = "8081"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,

		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
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
		"Public Profile: http://localhost:%s/api/v1/{handle}",
		port,
	)

	log.Printf(
		"OIDC issuer: %s",
		oidcIssuer,
	)

	log.Println("==============================================")

	// ============================================================
	// START SERVER
	// ============================================================

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}
}

// ============================================================
// HEALTH HANDLER
// ============================================================

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(
		"Content-Type",
		"text/plain; charset=utf-8",
	)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte("OK"))
}
