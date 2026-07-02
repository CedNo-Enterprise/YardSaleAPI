package server

import (
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/memory"
)

type AppServer struct {
	userRepository   user.UserRepository
	saleRepository   sale.SaleRepository
	sellerRepository seller.SellerRepository
}

func NewAppServer() *AppServer {
	return &AppServer{
		userRepository:   new(memory.InMemoryUserRepository),
		saleRepository:   new(memory.InMemorySaleRepository),
		sellerRepository: new(memory.InMemorySellerRepository),
	}
}

func (server *AppServer) GetUserRepository() *user.UserRepository {
	return &server.userRepository
}

func (server *AppServer) GetSaleRepository() *sale.SaleRepository {
	return &server.saleRepository
}

func (server *AppServer) GetSellerRepository() *seller.SellerRepository {
	return &server.sellerRepository
}
