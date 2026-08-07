package main

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/controllers"
	"net/http"
	"os"
	"time"
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
	s := server.NewAppServer()

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
