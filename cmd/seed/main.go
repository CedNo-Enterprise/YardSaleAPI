// Command seed fills the database with a known set of users, sales and
// itineraries for manual end-to-end testing.
//
// Run it from the repository root, with the database up:
//
//	go run ./cmd/seed
//
// Every seeded row uses a fixed id, so the command is safe to re-run: it clears
// its own previous rows first and leaves anything else in the database alone.
package main

import (
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/domain/buyer"
	"GarageSaleAPI/domain/itinerary"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/database"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Fixed ids so the seed is reproducible and the values can be saved in Postman.
const (
	buyerId         = "11111111-1111-4111-8111-111111111111"
	vendorId        = "22222222-2222-4222-8222-222222222222"
	sellerId        = "33333333-3333-4333-8333-333333333333"
	plannedRouteId  = "44444444-4444-4444-8444-444444444444"
	emptyRouteId    = "55555555-5555-4555-8555-555555555555"
	buyerProfileId  = "66666666-6666-4666-8666-666666666666"
	seedPassword    = "correcthorsebattery"
	buyerEmail      = "buyer@example.com"
	vendorEmail     = "vendor@example.com"
	buyerName       = "Sam the Saturday Shopper"
	bcryptSeedCost  = 14
	saleIdPrefix    = "aaaaaaaa-0000-4000-8000-00000000000"
	routeStartLat   = 45.4215
	routeStartLong  = -75.6972
	buyerHomeLine1  = "100 Bank St"
	buyerHomeCity   = "Ottawa"
	buyerPostalCode = "K1P 5N2"
)

type seedSale struct {
	id        string
	name      string
	line1     string
	city      string
	latitude  float64
	longitude float64
	// dayOffset and status spread the seeded sales out so every GET /sale filter
	// has something to match and something to exclude.
	dayOffset int
	status    sale.Status
}

// The ungeocoded entry is deliberate: nothing in the application populates
// coordinates yet, and an unlocatable stop sorts to the end of a route.
var seedSales = []seedSale{
	{saleIdPrefix + "1", "ByWard Market moving sale", "55 ByWard Market Sq", "Ottawa", 45.4285, -75.6920, 2, sale.StatusScheduled},
	{saleIdPrefix + "2", "Parliament Hill estate sale", "111 Wellington St", "Ottawa", 45.4236, -75.7009, 7, sale.StatusScheduled},
	{saleIdPrefix + "3", "Glebe multi-family sale", "180 Fifth Ave", "Ottawa", 45.4005, -75.6900, 7, sale.StatusScheduled},
	{saleIdPrefix + "4", "Westboro garage clear-out", "380 Richmond Rd", "Ottawa", 45.3900, -75.7530, 14, sale.StatusScheduled},
	{saleIdPrefix + "5", "Orleans driveway sale", "250 Centrum Blvd", "Orleans", 45.4600, -75.5200, 0, sale.StatusActive},
	{saleIdPrefix + "6", "Kanata tool sale", "145 Terry Fox Dr", "Kanata", 45.3100, -75.9000, -7, sale.StatusCompleted},
	{saleIdPrefix + "7", "Ungeocoded yard sale", "1 Unknown Rd", "Ottawa", 0, 0, 21, sale.StatusScheduled},
	{saleIdPrefix + "8", "Hintonburg sale (cancelled)", "1000 Wellington St W", "Ottawa", 45.4020, -75.7230, 9, sale.StatusCancelled},
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}

	db, err := database.NewGormDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err = database.AutoMigrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	ctx := context.Background()

	if err = clearPreviousSeed(db); err != nil {
		log.Fatalf("failed to clear previous seed data: %v", err)
	}

	if err = seedUsers(ctx, db); err != nil {
		log.Fatalf("failed to seed users: %v", err)
	}
	if err = seedSeller(ctx, db); err != nil {
		log.Fatalf("failed to seed seller: %v", err)
	}
	if err = seedBuyer(ctx, db); err != nil {
		log.Fatalf("failed to seed buyer: %v", err)
	}
	if err = seedSaleRows(ctx, db); err != nil {
		log.Fatalf("failed to seed sales: %v", err)
	}
	if err = seedItineraries(ctx, db); err != nil {
		log.Fatalf("failed to seed itineraries: %v", err)
	}

	printSummary()
}

func clearPreviousSeed(db *gorm.DB) error {
	saleIds := make([]string, 0, len(seedSales))
	for _, s := range seedSales {
		saleIds = append(saleIds, s.id)
	}
	routeIds := []string{plannedRouteId, emptyRouteId}
	userIds := []string{buyerId, vendorId}

	return db.Transaction(func(tx *gorm.DB) error {
		// sales.address_id and buyers.home_address_id are real foreign keys, so
		// the address ids have to be read out before the rows naming them go.
		var addressIds []int64
		if err := tx.Raw("SELECT address_id FROM sales WHERE id IN ?", saleIds).
			Scan(&addressIds).Error; err != nil {
			return err
		}

		var buyerAddressIds []int64
		if err := tx.Raw("SELECT home_address_id FROM buyers WHERE id = ? AND home_address_id IS NOT NULL", buyerProfileId).
			Scan(&buyerAddressIds).Error; err != nil {
			return err
		}
		addressIds = append(addressIds, buyerAddressIds...)

		statements := []struct {
			query string
			args  []any
		}{
			{"DELETE FROM itinerary_stops WHERE itinerary_id IN ?", []any{routeIds}},
			{"DELETE FROM itineraries WHERE id IN ?", []any{routeIds}},
			{"DELETE FROM sales WHERE id IN ?", []any{saleIds}},
			{"DELETE FROM buyers WHERE id = ?", []any{buyerProfileId}},
			{"DELETE FROM sellers WHERE id = ?", []any{sellerId}},
			{"DELETE FROM users WHERE id IN ?", []any{userIds}},
		}

		for _, statement := range statements {
			if err := tx.Exec(statement.query, statement.args...).Error; err != nil {
				return err
			}
		}

		if len(addressIds) == 0 {
			return nil
		}

		return tx.Exec("DELETE FROM addresses WHERE id IN ?", addressIds).Error
	})
}

func seedUsers(ctx context.Context, db *gorm.DB) error {
	repository := database.NewUserRepository(db)

	// Matches the cost UserService.AddUser uses, so these hashes verify on login.
	hash, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcryptSeedCost)
	if err != nil {
		return err
	}

	now := time.Now()
	users := []*user.User{
		user.CreateUser(buyerId, "seedbuyer", string(hash), buyerEmail, now),
		user.CreateUser(vendorId, "seedvendor", string(hash), vendorEmail, now),
	}

	for _, u := range users {
		if err = repository.Create(ctx, u); err != nil {
			return err
		}
	}

	return nil
}

func seedSeller(ctx context.Context, db *gorm.DB) error {
	repository := database.NewSellerRepository(db)
	s := seller.CreateSeller(sellerId, vendorId, "seedvendor", time.Now())

	return repository.Create(ctx, s)
}

func seedBuyer(ctx context.Context, db *gorm.DB) error {
	repository := database.NewBuyerRepository(db)

	b := buyer.CreateBuyer(buyerProfileId, buyerId, buyerName, time.Now())
	b.SetHomeAddress(address.CreateAddress(buyerHomeLine1, nil, buyerHomeCity, "ON", buyerPostalCode, "CA"))

	return repository.Create(ctx, b)
}

func seedSaleRows(ctx context.Context, db *gorm.DB) error {
	repository := database.NewSaleRepository(db)

	for _, seeded := range seedSales {
		saleAddress := address.CreateAddress(seeded.line1, nil, seeded.city, "ON", "K1A 0B1", "CA")
		// The API cannot set coordinates yet, so the seeder does it directly;
		// without them every route would fall back to submitted order.
		if seeded.latitude != 0 || seeded.longitude != 0 {
			saleAddress.AddLatLong(seeded.latitude, seeded.longitude)
		}

		// Hydrated rather than created, because CreateSale derives the status from
		// the date and these rows need statuses chosen for the search filters.
		s := sale.HydrateSale(
			seeded.id, sellerId, seeded.name, saleAddress, seedSaleDate(seeded),
			"Seeded for manual testing.", []sale.SaleItem{}, seeded.status, time.Now(),
		)
		if err := repository.Create(ctx, s); err != nil {
			return err
		}
	}

	return nil
}

func seedSaleDate(seeded seedSale) time.Time {
	return time.Now().AddDate(0, 0, seeded.dayOffset)
}

func seedItineraries(ctx context.Context, db *gorm.DB) error {
	repository := database.NewItineraryRepository(db)

	coordinates := make(map[string]itinerary.GeoPoint, len(seedSales))
	for _, seeded := range seedSales {
		coordinates[seeded.id] = itinerary.GeoPoint{Latitude: seeded.latitude, Longitude: seeded.longitude}
	}

	start := itinerary.GeoPoint{Latitude: routeStartLat, Longitude: routeStartLong}
	planned := itinerary.CreateItinerary(
		plannedRouteId, buyerId, "Saturday morning run", "Downtown and west",
		time.Now().AddDate(0, 0, 7), start.Latitude, start.Longitude, time.Now(),
	)
	for _, seeded := range seedSales {
		planned.AppendStop(seeded.id)
	}
	planned.ReorderFrom(start, coordinates)
	// One stop already visited, so status round-trips are visible immediately.
	planned.SetStopStatus(seedSales[0].id, itinerary.StopStatusVisited)

	if err := repository.Create(ctx, planned); err != nil {
		return err
	}

	empty := itinerary.CreateItinerary(
		emptyRouteId, buyerId, "Next weekend (empty)", "Nothing planned yet",
		time.Now().AddDate(0, 0, 14), 0, 0, time.Now(),
	)

	return repository.Create(ctx, empty)
}

func printSummary() {
	fmt.Println()
	fmt.Println("Seed complete.")
	fmt.Println()
	fmt.Println("Login (POST /login) — both users share this password:")
	fmt.Printf("  password: %s\n", seedPassword)
	fmt.Printf("  buyer    %s   (owns both itineraries, has a buyer profile)\n", buyerEmail)
	fmt.Printf("  vendor   %s   (owns the sales; use to check 403s)\n", vendorEmail)
	fmt.Println()
	fmt.Println("Ids:")
	fmt.Printf("  buyer user        %s\n", buyerId)
	fmt.Printf("  buyer profile     %s\n", buyerProfileId)
	fmt.Printf("  vendor user       %s\n", vendorId)
	fmt.Printf("  seller            %s\n", sellerId)
	fmt.Printf("  itinerary (full)  %s\n", plannedRouteId)
	fmt.Printf("  itinerary (empty) %s\n", emptyRouteId)
	fmt.Println()
	fmt.Println("Sales:")
	for _, seeded := range seedSales {
		location := fmt.Sprintf("%.4f, %.4f", seeded.latitude, seeded.longitude)
		if seeded.latitude == 0 && seeded.longitude == 0 {
			location = "not geocoded"
		}
		fmt.Printf(
			"  %s  %-30s %-10s %s  %s\n",
			seeded.id, seeded.name, seeded.status,
			seedSaleDate(seeded).Format(time.RFC3339), location,
		)
	}
	fmt.Println()
	fmt.Printf("Route start point: %.4f, %.4f\n", routeStartLat, routeStartLong)
	fmt.Println()
	fmt.Println("Try: GET /sale, GET /sale?status=cancelled, GET /sale?limit=2&offset=2")
}
