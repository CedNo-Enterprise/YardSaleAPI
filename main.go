package main

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/infrastructure/persistence/database"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/controllers"
	"log"
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

func initAppState(mux *http.ServeMux) {
	loadEnv()
	db := setupDatabase()
	s := server.NewAppServer(db)

	jwtKey := []byte(os.Getenv("JWT_SECRET"))

	tokenService := services.NewTokenService(jwtKey, 24*time.Hour)

	authMiddleware := interfaces.NewAuthenticationMiddleware(tokenService)

	userService := services.NewUserService(*s.GetUserRepository(), tokenService)
	userController := controllers.NewUserController(userService)
	userController.AddUserHandlersToMux(mux)

	saleService := services.NewSaleService(*s.GetSaleRepository())
	saleController := controllers.NewSaleController(saleService, authMiddleware)
	saleController.AddSalesHandlersToMux(mux)

	sellerService := services.NewSellerService(*s.GetSellerRepository(), *s.GetUserRepository())
	sellerController := controllers.NewSellerController(sellerService, authMiddleware)
	sellerController.AddSalesHandlersToMux(mux)
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
