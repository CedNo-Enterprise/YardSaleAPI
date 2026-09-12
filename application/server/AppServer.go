package server

import (
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/database"

	"gorm.io/gorm"
)

type AppServer struct {
	userRepository   user.UserRepository
	saleRepository   sale.SaleRepository
	sellerRepository seller.SellerRepository
}

func NewAppServer(db *gorm.DB) *AppServer {
	return &AppServer{
		userRepository:   database.NewUserRepository(db),
		saleRepository:   database.NewSaleRepository(db),
		sellerRepository: database.NewSellerRepository(db),
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
