package main

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/domain/loginattempt"
	"GarageSaleAPI/domain/token"
	"GarageSaleAPI/infrastructure/persistence/database"
	"GarageSaleAPI/infrastructure/ratelimit"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/controllers"
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	mux := http.NewServeMux()
	initAppState(mux)

	err := http.ListenAndServe("localhost:8080", mux)

	if err != nil {
		return
	}
}

// Login throttling budgets. Per-username catches one account ground from many
// addresses; per-IP catches one address spraying many accounts.
var (
	usernameLoginPolicy = services.Policy{Limit: 10, Window: 15 * time.Minute}
	ipLoginPolicy       = services.Policy{Limit: 30, Window: 15 * time.Minute}
)

func initAppState(mux *http.ServeMux) {
	loadEnv()
	db := setupDatabase()
	s := server.NewAppServer(db)

	jwtKey := []byte(os.Getenv("JWT_SECRET"))

	tokenService := services.NewTokenService(jwtKey, 24*time.Hour)

	revokedTokens := *s.GetRevokedTokenRepository()
	sessionService := services.NewSessionService(revokedTokens)

	authMiddleware := interfaces.NewAuthenticationMiddleware(tokenService, sessionService)

	attemptStore := ratelimit.NewInMemoryAttemptStore(ratelimit.DefaultMaxEntries)
	loginThrottle := services.NewLoginThrottleService(attemptStore, usernameLoginPolicy, ipLoginPolicy)

	userService := services.NewUserService(*s.GetUserRepository(), tokenService)
	userController := controllers.NewUserController(userService, sessionService, loginThrottle, authMiddleware)
	userController.AddUserHandlersToMux(mux)

	saleService := services.NewSaleService(*s.GetSaleRepository())
	saleController := controllers.NewSaleController(saleService, authMiddleware)
	saleController.AddSalesHandlersToMux(mux)

	sellerService := services.NewSellerService(*s.GetSellerRepository(), *s.GetUserRepository())
	sellerController := controllers.NewSellerController(sellerService, authMiddleware)
	sellerController.AddSalesHandlersToMux(mux)

	startRevokedTokenCleanup(revokedTokens, time.Hour)
	startLoginAttemptCleanup(attemptStore, 5*time.Minute)
}

// startRevokedTokenCleanup periodically drops denylist rows for tokens that
// have expired on their own, so the table stays bounded by the token TTL.
func startRevokedTokenCleanup(revokedTokens token.RevokedTokenRepository, every time.Duration) {
	sweep := func() {
		if err := revokedTokens.DeleteExpired(context.Background(), time.Now()); err != nil {
			slog.Error("error cleaning up expired revoked tokens", "err", err.Error())
		}
	}

	sweep()

	go func() {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for range ticker.C {
			sweep()
		}
	}()
}

// startLoginAttemptCleanup drops throttling state whose failure log and block
// have both lapsed, keeping the store from accumulating one-off callers.
func startLoginAttemptCleanup(attempts loginattempt.AttemptStore, every time.Duration) {
	sweep := func() {
		attempts.DeleteStale(time.Now())
	}

	sweep()

	go func() {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for range ticker.C {
			sweep()
		}
	}()
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}
}

func setupDatabase() *gorm.DB {
	gormDB, err := database.NewGormDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err = database.AutoMigrate(gormDB); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	return gormDB
}
