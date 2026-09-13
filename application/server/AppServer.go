package server

import (
	"GarageSaleAPI/domain/buyer"
	"GarageSaleAPI/domain/itinerary"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/token"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/database"

	"gorm.io/gorm"
)

type AppServer struct {
	userRepository         user.UserRepository
	saleRepository         sale.SaleRepository
	sellerRepository       seller.SellerRepository
	buyerRepository        buyer.BuyerRepository
	revokedTokenRepository token.RevokedTokenRepository
	itineraryRepository    itinerary.ItineraryRepository
}

func NewAppServer(db *gorm.DB) *AppServer {
	return &AppServer{
		userRepository:         database.NewUserRepository(db),
		saleRepository:         database.NewSaleRepository(db),
		sellerRepository:       database.NewSellerRepository(db),
		buyerRepository:        database.NewBuyerRepository(db),
		revokedTokenRepository: database.NewRevokedTokenRepository(db),
		itineraryRepository:    database.NewItineraryRepository(db),
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

func (server *AppServer) GetBuyerRepository() *buyer.BuyerRepository {
	return &server.buyerRepository
}

func (server *AppServer) GetRevokedTokenRepository() *token.RevokedTokenRepository {
	return &server.revokedTokenRepository
}

func (server *AppServer) GetItineraryRepository() *itinerary.ItineraryRepository {
	return &server.itineraryRepository
}
