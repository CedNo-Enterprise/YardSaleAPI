package main

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/interfaces/controllers"
	"net/http"
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

	userService := services.NewUserService(*s.GetUserRepository())
	userController := controllers.NewUserController(userService)
	userController.AddUserHandlersToMux(mux)

	saleService := services.NewSaleService(*s.GetSaleRepository())
	saleController := controllers.NewSaleController(saleService)
	saleController.AddSalesHandlersToMux(mux)
}
