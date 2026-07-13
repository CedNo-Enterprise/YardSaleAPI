package controllers

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/services"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/test"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSaleController_addSale(t *testing.T) {
	s := server.NewAppServer()
	controller := *NewSaleController(services.NewSaleService(*s.GetSaleRepository()))

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
			name: "Add valid user",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequest(
					"POST",
					"/sale",
					bytes.NewBufferString(`{
						"SellerId": "sellerId",
						"Name": "New Sale on the Block!",
    					"Address": {"line1":"northern","city":"Washington","state":"WS","postal_code":"U1A 2C5","country":"US"},
						"Date": "2026-07-06T19:28:00Z"
					}`),
					"application/json"),
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "Add valid user",
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
			},
			wantStatusCode: http.StatusUnsupportedMediaType,
		},
		{
			name: "Add invalid user",
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
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Cleanup(func() {
			s = server.NewAppServer()
		})

		t.Run(tt.name, func(t *testing.T) {
			controller.addSale(tt.args.w, tt.args.r)

			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("addSale() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestSaleController_getSale(t *testing.T) {
	s := server.NewAppServer()
	service := services.NewSaleService(*s.GetSaleRepository())
	controller := *NewSaleController(service)

	saleToAdd := requests.SaleRequest{
		SellerId: uuid.NewString(),
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
	saleId, err := service.AddSale(ctx, saleToAdd)
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
			name: "Get valid sale",
			args: args{
				w: httptest.NewRecorder(),
				r: test.CreateRequestWithPathParam(http.MethodGet, "/sale/", nil, "id", *saleId),
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"name":"Best sale in the east","address":{"line1":"northern","city":"Washington","state":"WS","postal_code":"U1A 2C5","country":"US","latitude":0,"longitude":0}}` + "\n",
		},
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
			t.Cleanup(func() {
				s = server.NewAppServer()
			})

			controller.getSale(tt.args.w, tt.args.r)

			if tt.wantStatusCode != tt.args.w.Code {
				t.Errorf("getSale() got status code = %v, want = %v", tt.args.w.Code, tt.wantStatusCode)
			}
			if tt.wantBody != tt.args.w.Body.String() {
				t.Errorf("getSale() got body = %v, want = %v", tt.args.w.Body.String(), tt.wantBody)
			}
		})
	}
}
