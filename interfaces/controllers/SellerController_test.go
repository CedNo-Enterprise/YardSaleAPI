package controllers

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/test"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSellerController_addSeller(t *testing.T) {
	s := server.NewAppServer()
	userRepo := *s.GetUserRepository()
	sellerService := services.NewSellerService(*s.GetSellerRepository(), userRepo)
	_ = userRepo.Save(
		context.Background(),
		user.CreateUser("username", "password", "email@email.com", time.Now()),
	)

	controller := NewSellerController(sellerService)

	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
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
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "Add invalid seller",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					"POST",
					"/seller",
					bytes.NewBufferString(`{
						"username":        "invalid_username"
					}`),
					"application/json"),
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Cleanup(func() {
			s = server.NewAppServer()
		})

		t.Run(tt.name, func(t *testing.T) {
			controller.addSeller(tt.args.w, tt.args.r)

			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("addSeller() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestSellerController_getSellerById(t *testing.T) {
	s := server.NewAppServer()
	userRepo := *s.GetUserRepository()
	sellerRepo := *s.GetSellerRepository()
	sellerService := services.NewSellerService(sellerRepo, userRepo)
	_ = userRepo.Save(
		context.Background(),
		user.CreateUser("username", "password", "email@email.com", time.Now()),
	)
	addedSeller := seller.CreateSeller("seller_id", "username", time.Now())
	_ = sellerRepo.Save(context.Background(), addedSeller)

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
			wantBody:       `{"id":"seller_id","username":"username","saved_addresses":[],"inventory":[]}` + "\n",
		},
		{
			name: "Get non-added seller by id",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequestWithPathParam("POST", "/seller/", nil, "id", "invalid_seller_id"),
			},
			wantStatusCode: http.StatusBadRequest,
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

func TestSellerController_getSellerByUsername(t *testing.T) {
	s := server.NewAppServer()
	userRepo := *s.GetUserRepository()
	sellerRepo := *s.GetSellerRepository()
	sellerService := services.NewSellerService(sellerRepo, userRepo)
	_ = userRepo.Save(
		context.Background(),
		user.CreateUser("username", "password", "email@email.com", time.Now()),
	)
	addedSeller := seller.CreateSeller("seller_id", "username", time.Now())
	_ = sellerRepo.Save(context.Background(), addedSeller)

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
				r: test.CreateRequestWithPathParam("POST", "/seller/", nil, "username", "username"),
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"id":"seller_id","username":"username","saved_addresses":[],"inventory":[]}` + "\n",
		},
		{
			name: "Get non-added seller by id",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequestWithPathParam("POST", "/seller/", nil, "username", "invalid_username"),
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := &SellerController{
				sellerService: sellerService,
			}
			controller.getSellerByUsername(tt.args.w, tt.args.r)

			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("getSellerByUsername() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
			if tt.wantBody != tt.args.w.Body.String() {
				t.Errorf("getSellerByUsername() got body = %v, want = %v", tt.args.w.Body, tt.wantBody)
			}
		})
	}
}
