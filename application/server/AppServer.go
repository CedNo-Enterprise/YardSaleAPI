package server

import (
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/memory"
)

type AppServer struct {
	userRepository user.UserRepository
	saleRepository sale.SaleRepository
}

func NewAppServer() *AppServer {
	return &AppServer{
		userRepository: new(memory.InMemoryUserRepository),
		saleRepository: new(memory.InMemorySaleRepository),
	}
}

func (server *AppServer) GetUserRepository() *user.UserRepository {
	return &server.userRepository
}

func (server *AppServer) GetSaleRepository() *sale.SaleRepository {
	return &server.saleRepository
}
