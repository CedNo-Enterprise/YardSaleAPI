package main

import (
	"GarageSaleAPI/interfaces/controllers"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	controllers.AddUserHandlersToMux(mux)

	err := http.ListenAndServe("localhost:8080", mux)

	if err != nil {
		return
	}
}
