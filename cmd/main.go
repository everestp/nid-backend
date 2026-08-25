package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
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

	socialCtrl "nid-backend/modules/social/controller"
	socialRepo "nid-backend/modules/social/repository"
	socialSvc "nid-backend/modules/social/service"

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

	// Social
	socialRepository := socialRepo.NewSocialRepository(db)

	// ============================================================
	// OIDC Signing Key
	// ============================================================

	oidcPrivateKey, err := loadOIDCPrivateKey()
	if err != nil {
		log.Fatalf("failed to load OIDC private key: %v", err)
	}

	// ============================================================
	// Services
	// ============================================================

	authService := authSvc.NewAuthService(
		authRepository,
	)

	handleService := handleSvc.NewHandleService(
		handleRepository,
	)

	walletService := walletSvc.NewWalletService(
		walletRepository,
	)

	resolutionService := resSvc.NewResolutionService(
		resolutionRepository,
	)

	sessionService := sesSvc.NewSessionService(
		sessionRepository,
	)

	userService := userSvc.NewUserService(
		userRepository,
	)

	// Social
	socialService := socialSvc.NewSocialService(
		socialRepository,
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

	oidcService := oidcSvc.NewOIDCService(
		oidcRepository,
		oidcPrivateKey,
		oidcIssuer,
		oidcKeyID,
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

	// Social
	socialController := socialCtrl.NewSocialController(
		socialService,
	)

	// ============================================================
	// Public Router
	// ============================================================

	mux := http.NewServeMux()

	// ============================================================
	// Health
	// ============================================================

	mux.HandleFunc(
		"GET /health",
		healthHandler,
	)

	// ============================================================
	// Core Public API
	// ============================================================

	mux.HandleFunc(
		"POST /api/v1/auth/login",
		authController.LoginHandler,
	)

	mux.HandleFunc(
		"GET /api/v1/resolve",
		resolutionController.ResolveHandler,
	)

	mux.HandleFunc(
		"POST /api/v1/handles/claim",
		handleController.ClaimHandler,
	)

	// ============================================================
	// OAuth 2.0 / OpenID Connect
	// ============================================================

	mux.HandleFunc(
		"POST /oauth/register",
		oidcController.RegisterClientHandler,
	)

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

	// ============================================================
	// OpenID Connect Discovery
	// ============================================================

	mux.HandleFunc(
		"GET /.well-known/openid-configuration",
		oidcController.DiscoveryHandler,
	)

	// ============================================================
	// JWKS
	// ============================================================

	mux.HandleFunc(
		"GET /.well-known/jwks.json",
		oidcController.JWKSHandler,
	)

	// ============================================================
	// Protected Router
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
	// Session
	// ============================================================

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

protectedMux.HandleFunc(
	"GET /api/v1/social",
	socialController.ListHandler,
)

protectedMux.HandleFunc(
	"GET /api/v1/social/public",
	socialController.PublicListHandler,
)

protectedMux.HandleFunc(
	"GET /api/v1/social/{id}",
	socialController.GetHandler,
)

protectedMux.HandleFunc(
	"POST /api/v1/social",
	socialController.CreateHandler,
)

protectedMux.HandleFunc(
	"PUT /api/v1/social/{id}",
	socialController.UpdateHandler,
)

protectedMux.HandleFunc(
	"PATCH /api/v1/social/{id}/visibility",
	socialController.ToggleVisibilityHandler,
)

protectedMux.HandleFunc(
	"DELETE /api/v1/social/{id}",
	socialController.DeleteHandler,
)

	// ============================================================
	// Protected API Mounts
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

	mux.Handle(
		"/api/v1/social",
		middleware.AuthMiddleware(protectedMux),
	)

	// ============================================================
	// Global Middleware
	// ============================================================

	handler := middleware.CORSMiddleware(
		config.RequestLogger(mux),
	)

	// ============================================================
	// HTTP Server
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

	log.Printf("NID backend starting on port %s", port)

	log.Printf(
		"OIDC issuer: %s",
		oidcIssuer,
	)

	log.Printf(
		"OIDC authorization: %s/oauth/authorize",
		oidcIssuer,
	)

	log.Printf(
		"OIDC approval: %s/oauth/authorize/approve",
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

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}
}

// ============================================================
// Health Handler
// ============================================================

func healthHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(
		"Content-Type",
		"text/plain; charset=utf-8",
	)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte("OK"))
}

// ============================================================
// Environment Helper
// ============================================================

func getEnv(key, fallback string) string {

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
		return nil, errors.New(
			"invalid RSA private key PEM",
		)
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
		return nil, errors.New(
			"invalid PKCS#8 private key",
		)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)

	if !ok {
		return nil, errors.New(
		"OIDC private key is not RSA",
		)
	}

	return rsaKey, nil
}
