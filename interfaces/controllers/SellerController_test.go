package controllers

import (
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces"
	"GarageSaleAPI/test"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSellerController_addSeller(t *testing.T) {
	userRepo := &memory.InMemoryUserRepository{}
	tokenService := services.NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)
	sellerService := services.NewSellerService(&memory.InMemorySellerRepository{}, userRepo)
	userId := uuid.NewString()
	_ = userRepo.Create(
		context.Background(),
		user.CreateUser(userId, "username", "password", "email@email.com", time.Now()),
	)

	controller := NewSellerController(sellerService, interfaces.NewAuthenticationMiddleware(tokenService))

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
			name: "Add valid seller",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					"POST",
					"/seller",
					bytes.NewBufferString(`{
						"username":        "username"
					}`),
					"application/json"),
				userId: userId,
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "Add seller with invalid user id",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					"POST",
					"/seller",
					bytes.NewBufferString(`{
						"username":        "invalid_username"
					}`),
					"application/json"),
				userId: uuid.NewString(),
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller.addSeller(tt.args.w, tt.args.r, tt.args.userId)

			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("addSeller() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestSellerController_getSellerById(t *testing.T) {
	sellerService := setupGetSellerTests()

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
			name: "Get added seller by id",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequestWithPathParam("POST", "/seller/", nil, "id", "seller_id"),
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"id":"seller_id","name":"username","saved_addresses":[],"inventory":[]}` + "\n",
		},
		{
			name: "Get non-added seller by id",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequestWithPathParam("POST", "/seller/", nil, "id", "invalid_seller_id"),
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "seller not found" + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := &SellerController{
				sellerService: sellerService,
			}
			controller.getSellerById(tt.args.w, tt.args.r)

			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("getSellerById() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
			if tt.wantBody != tt.args.w.Body.String() {
				t.Errorf("getSellerById() got body = %v, want = %v", tt.args.w.Body, tt.wantBody)
			}
		})
	}
}

func TestSellerController_getSellerByUserId(t *testing.T) {
	sellerService := setupGetSellerTests()

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
			name: "Get added seller by user id",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequestWithPathParam("POST", "/seller/", nil, "userId", "username"),
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"id":"seller_id","name":"username","saved_addresses":[],"inventory":[]}` + "\n",
		},
		{
			name: "Get non-added seller by user id",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequestWithPathParam("POST", "/seller/", nil, "userId", "invalid_user_id"),
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "seller not found" + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := &SellerController{
				sellerService: sellerService,
			}
			controller.getSellerByUserId(tt.args.w, tt.args.r)

			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("getSellerByUserId() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
			if tt.wantBody != tt.args.w.Body.String() {
				t.Errorf("getSellerByUserId() got body = %v, want = %v", tt.args.w.Body, tt.wantBody)
			}
		})
	}
}

func setupGetSellerTests() *services.SellerService {
	userRepo := &memory.InMemoryUserRepository{}
	sellerRepo := &memory.InMemorySellerRepository{}
	sellerService := services.NewSellerService(sellerRepo, userRepo)
	_ = userRepo.Create(
		context.Background(),
		user.CreateUser(uuid.NewString(), "username", "password", "email@email.com", time.Now()),
	)
	addedSeller := seller.CreateSeller("seller_id", "username", "username", time.Now())
	_ = sellerRepo.Create(context.Background(), addedSeller)

	return sellerService
}
