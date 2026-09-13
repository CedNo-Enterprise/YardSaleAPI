package controllers

import (
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/interfaces/responses"
	"GarageSaleAPI/test"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	saleSellerId      = "11111111-1111-4111-8111-111111111111"
	saleSellerUserId  = "22222222-2222-4222-8222-222222222222"
	saleStrangerUsrId = "33333333-3333-4333-8333-333333333333"
)

func newSaleControllerWithSeller(t *testing.T) (*SaleController, *services.SaleService) {
	t.Helper()

	sellerRepo := &memory.InMemorySellerRepository{}
	owner := seller.CreateSeller(saleSellerId, saleSellerUserId, "seedvendor", time.Now())
	if err := sellerRepo.Create(test.CreateTestContext(t), owner); err != nil {
		t.Fatalf("seeding seller: %v", err)
	}

	service := services.NewSaleService(&memory.InMemorySaleRepository{}, sellerRepo)
	tokenService := services.NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)
	authMiddleware := interfaces.NewAuthenticationMiddleware(
		tokenService, services.NewSessionService(&memory.InMemoryRevokedTokenRepository{}),
	)

	return NewSaleController(service, authMiddleware), service
}

func TestSaleController_addSale(t *testing.T) {
	controller, _ := newSaleControllerWithSeller(t)

	type args struct {
		w      *httptest.ResponseRecorder
		r      *http.Request
		userId string
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
	}{
		{
			name: "Add valid sale",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					"POST",
					"/sale",
					bytes.NewBufferString(`{
						"SellerId": "11111111-1111-4111-8111-111111111111",
						"Name": "New Sale on the Block!",
    					"Address": {"line1":"northern","city":"Washington","state":"WS","postal_code":"U1A 2C5","country":"US"},
						"Date": "2026-07-06T19:28:00Z"
					}`),
					"application/json"),
				userId: saleSellerUserId,
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "Add sale under a seller the caller does not own",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					"POST",
					"/sale",
					bytes.NewBufferString(`{
						"SellerId": "11111111-1111-4111-8111-111111111111",
						"Name": "Sale under someone else's name",
    					"Address": {"line1":"northern","city":"Washington","state":"WS","postal_code":"U1A 2C5","country":"US"},
						"Date": "2026-07-06T19:28:00Z"
					}`),
					"application/json"),
				userId: saleStrangerUsrId,
			},
			wantStatusCode: http.StatusForbidden,
		},
		{
			name: "Add sale under a seller that does not exist",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					"POST",
					"/sale",
					bytes.NewBufferString(`{
						"SellerId": "99999999-9999-4999-8999-999999999999",
						"Name": "Sale with no seller",
    					"Address": {"line1":"northern","city":"Washington","state":"WS","postal_code":"U1A 2C5","country":"US"},
						"Date": "2026-07-06T19:28:00Z"
					}`),
					"application/json"),
				userId: saleSellerUserId,
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "Add sale with wrong content type",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					"POST",
					"/sale",
					bytes.NewBufferString(`{
						"Name": "New Sale on the Block!",
    					"Address": "123 st road"
					}`),
					""),
				userId: "Edgouille",
			},
			wantStatusCode: http.StatusUnsupportedMediaType,
		},
		{
			name: "Add invalid sale",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					"POST",
					"/sale",
					bytes.NewBufferString(`{
						"Name": "Sale",
    					"Address": "123 st road"
					}`),
					"application/json"),
				userId: "Edgouille",
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller.addSale(tt.args.w, tt.args.r, tt.args.userId)

			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("addSale() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestSaleController_getSale(t *testing.T) {
	controller, service := newSaleControllerWithSeller(t)

	saleToAdd := requests.SaleRequest{
		SellerId: saleSellerId,
		Name:     "Best sale in the east",
		Address: requests.AddressRequest{
			Line1:      "northern",
			Line2:      "",
			City:       "Washington",
			State:      "WS",
			PostalCode: "U1A 2C5",
			Country:    "US",
		},
		Date: time.Now(),
	}
	ctx := test.CreateTestContext(t)
	saleId, err := service.AddSale(ctx, saleSellerUserId, saleToAdd)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "Get nonexistent sale",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequestWithPathParam(http.MethodGet, "/sale/", nil, "id", "invalid saleId"),
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "sale not found\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller.getSale(tt.args.w, tt.args.r)

			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("getSale() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
			if tt.wantBody != tt.args.w.Body.String() {
				t.Errorf("getSale() got body = %v, want = %v", tt.args.w.Body.String(), tt.wantBody)
			}
		})
	}

	t.Run("Get valid sale", func(t *testing.T) {
		w := httptest.NewRecorder()

		controller.getSale(w, test.CreateRequestWithPathParam(http.MethodGet, "/sale/", nil, "id", *saleId))

		if w.Code != http.StatusOK {
			t.Fatalf("getSale() got status code = %v, want = %v", w.Code, http.StatusOK)
		}

		var response responses.SaleResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("decoding body %q: %v", w.Body.String(), err)
		}

		if response.Id != *saleId {
			t.Errorf("getSale() id = %q, want %q", response.Id, *saleId)
		}
		if response.SellerId != saleToAdd.SellerId {
			t.Errorf("getSale() seller_id = %q, want %q", response.SellerId, saleToAdd.SellerId)
		}
		if response.Name != "Best sale in the east" {
			t.Errorf("getSale() name = %q, want %q", response.Name, "Best sale in the east")
		}
		if response.Status != sale.StatusScheduled {
			t.Errorf("getSale() status = %q, want %q", response.Status, sale.StatusScheduled)
		}
		if response.Address.City != "Washington" {
			t.Errorf("getSale() address.city = %q, want %q", response.Address.City, "Washington")
		}
		if response.Items == nil {
			t.Errorf("getSale() items = null, want an empty list")
		}
	})
}

func TestSaleController_addSale_rejectedRequestWritesOneResponse(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		contentType    string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "wrong content type",
			body:           `{"sellerId":"s","name":"Sale","address":{},"date":"2026-07-06T19:28:00Z"}`,
			contentType:    "text/plain",
			wantStatusCode: http.StatusUnsupportedMediaType,
			wantBody:       "invalid content type\n",
		},
		{
			name:           "malformed body",
			body:           `{"name": `,
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "bad request body\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, _ := newSaleControllerWithSeller(t)

			w := httptest.NewRecorder()
			r := test.CreateRequest(http.MethodPost, "/sale", bytes.NewBufferString(tt.body), tt.contentType)

			controller.addSale(w, r, uuid.NewString())

			test.ValidateExpectedCodeAndBody(w, t, tt.wantStatusCode, tt.wantBody)
			if w.Header().Get("Location") != "" {
				t.Errorf("a rejected request set a Location header")
			}
		})
	}
}
