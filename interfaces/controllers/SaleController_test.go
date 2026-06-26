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
						"Name": "New Sale on the Block!",
    					"Address": "123 st road"
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
		})
	}
}

func TestSaleController_getSale(t *testing.T) {
	s := server.NewAppServer()
	service := services.NewSaleService(*s.GetSaleRepository())
	controller := *NewSaleController(service)

	saleToAdd := requests.SaleRequest{
		Name: "New Sale on the Block!",
		Address: requests.AddressRequest{
			Line1:      "northern",
			Line2:      "",
			City:       "Washington",
			State:      "WS",
			PostalCode: "U1A 2C5",
			Country:    "US",
		},
	}
	ctx := test.CreateTestContext(t)
	_, err := service.AddSale(ctx, saleToAdd)
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
				r: test.CreateRequestWithPathParam(http.MethodGet, "/sale/", nil, "username", "saleId"),
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{}`,
		},
		{
			name: "Get nonexistent sale",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/sale/10001", nil),
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "user not found\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				s = server.NewAppServer()
			})

			controller.getSale(tt.args.w, tt.args.r)
		})
	}
}
