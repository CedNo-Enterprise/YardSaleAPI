package services

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/test"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

var validAddressRequest = requests.AddressRequest{
	Line1:      "northern",
	Line2:      "",
	City:       "Washington",
	State:      "WS",
	PostalCode: "U1A 2C5",
	Country:    "US",
}

var validAddress = address.CreateAddress(
	"123e4567-e89b-12d3-a456-426614174111", "northern", nil,
	"Washington", "WS", "U1A 2C5", "US", time.Now(),
)

func TestSaleService_AddSale(t *testing.T) {
	s := server.NewAppServer()

	type args struct {
		service *SaleService
		saleDTO requests.SaleRequest
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrText string
	}{
		{
			name: "add valid sale",
			args: args{
				service: NewSaleService(*s.GetSaleRepository()),
				saleDTO: requests.SaleRequest{
					Name:    "Best sale in the east",
					Address: validAddressRequest,
				},
			},
			wantErr: false,
		},
		{
			name: "add invalid sale",
			args: args{
				service: NewSaleService(*s.GetSaleRepository()),
				saleDTO: requests.SaleRequest{
					Name:    "",
					Address: validAddressRequest,
				},
			},
			wantErr:     true,
			wantErrText: "invalid sale",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := test.CreateTestContext(t)
			t.Cleanup(func() {
				s = server.NewAppServer()
			})

			if _, err := tt.args.service.AddSale(ctx, tt.args.saleDTO); (err != nil) != tt.wantErr ||
				(err != nil) && err.Error() != tt.wantErrText {
				t.Errorf("AddSale() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaleService_GetSaleById(t *testing.T) {
	s := server.NewAppServer()
	repo := *s.GetSaleRepository()
	saleId := uuid.NewString()
	newSale := sale.CreateSale(saleId, "newSale", validAddress)

	type args struct {
		service *SaleService
		saleId  string
	}
	tests := []struct {
		name    string
		args    args
		want    *sale.Sale
		wantErr bool
	}{
		{
			name: "Get sale by id",
			args: args{
				service: NewSaleService(repo),
				saleId:  saleId,
			},
			want:    &newSale,
			wantErr: false,
		},
		{
			name: "Get nonexistent sale by id",
			args: args{
				service: NewSaleService(repo),
				saleId:  "123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := test.CreateTestContext(t)
			_ = repo.Save(ctx, newSale)
			got, err := tt.args.service.GetSaleById(ctx, tt.args.saleId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSaleById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSaleById() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateSale(t *testing.T) {
	type args struct {
		saleDTO requests.SaleRequest
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrText string
	}{
		{
			name: "valid sale",
			args: args{
				saleDTO: requests.SaleRequest{
					Name:    "Best sale in the east",
					Address: validAddressRequest,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid sale empty name",
			args: args{
				saleDTO: requests.SaleRequest{
					Name:    "",
					Address: validAddressRequest,
				},
			},
			wantErr:     true,
			wantErrText: "invalid sale",
		},
		{
			name: "invalid sale long name",
			args: args{
				saleDTO: requests.SaleRequest{
					Name:    "This sale name is way too long for our liking and will be deemed invalid",
					Address: validAddressRequest,
				},
			},
			wantErr:     true,
			wantErrText: "invalid sale",
		},
		{
			name: "invalid sale empty address",
			args: args{
				saleDTO: requests.SaleRequest{
					Name: "Best sale in the east",
				},
			},
			wantErr:     true,
			wantErrText: "invalid sale",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSale(tt.args.saleDTO); (err != nil) != tt.wantErr ||
				(err != nil) && err.Error() != tt.wantErrText {
				t.Errorf("validateSale() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
