// cmd/main.go
package main

import (
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
	// Load environment variables from .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	cfg := config.LoadConfig()
	db, err := database.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Repositories
	authRepoInstance := authRepo.NewAuthRepository(db)
	handleRepoInstance := handleRepo.NewHandleRepository(db)
	walletRepoInstance := walletRepo.NewWalletRepository(db)
	resolutionRepoInstance := resRepo.NewResolutionRepository(db)
	sessionRepoInstance := sesRepo.NewSessionRepository(db)
	userRepoInstance := userRepo.NewUserRepository(db)
	oidcRepoInstance := oidcRepo.NewOIDCRepository(db)

	// Initialize Services
	authService := authSvc.NewAuthService(authRepoInstance)
	handleService := handleSvc.NewHandleService(handleRepoInstance)
	walletService := walletSvc.NewWalletService(walletRepoInstance)
	resolutionService := resSvc.NewResolutionService(resolutionRepoInstance)
	sessionService := sesSvc.NewSessionService(sessionRepoInstance)
	userService := userSvc.NewUserService(userRepoInstance)
	oidcService := oidcSvc.NewOIDCService(oidcRepoInstance)

	// Initialize Controllers
	authController := authCtrl.NewAuthController(authService)
	handleController := handleCtrl.NewHandleController(handleService)
	walletController := walletCtrl.NewWalletController(walletService)
	resolutionController := resCtrl.NewResolutionController(resolutionService)
	sessionController := sesCtrl.NewSessionController(sessionService)
	userController := userCtrl.NewUserController(userService)
	oidcController := oidcCtrl.NewOIDCController(oidcService)

	mux := http.NewServeMux()

	// 1. Health Check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 2. Core Public Routes (Accessible without a Bearer Token)
	mux.HandleFunc("/api/v1/auth/login", authController.LoginHandler)
	mux.HandleFunc("/api/v1/resolve", resolutionController.ResolveHandler)
	mux.HandleFunc("/api/v1/handles/claim", handleController.ClaimHandler) // Public homepage handle registration

	// 3. OIDC / OAuth 2.0 Public Routes (For External Apps / Sign in with NID)
	mux.HandleFunc("/oauth/authorize", oidcController.AuthorizeHandler)
	mux.HandleFunc("/oauth/token", oidcController.TokenHandler)
	mux.HandleFunc("/.well-known/openid-configuration", oidcController.DiscoveryHandler)

	// 4. Protected Routes (Requires Bearer Token)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/api/v1/wallets/link", walletController.LinkWalletHandler)
	protectedMux.HandleFunc("/api/v1/sessions/revoke", sessionController.RevokeHandler)
	protectedMux.HandleFunc("/api/v1/user/profile", userController.GetProfileHandler)

	// Register Protected Routes with Auth Middleware
	mux.Handle("/api/v1/wallets/", middleware.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/sessions/", middleware.AuthMiddleware(protectedMux))
	mux.Handle("/api/v1/user/", middleware.AuthMiddleware(protectedMux))

	// Apply Global Middleware (CORS & Request Logging)
	handler := middleware.CORSMiddleware(config.RequestLogger(mux))

	log.Printf("Server starting on port %s...", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
