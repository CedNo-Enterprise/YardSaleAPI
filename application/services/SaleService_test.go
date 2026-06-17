package services

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/interfaces/requests"
	"reflect"
	"testing"

	"github.com/google/uuid"
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
					Address: "123 road st",
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
					Address: "123 road st",
				},
			},
			wantErr:     true,
			wantErrText: "invalid sale",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				s = server.NewAppServer()
			})

			if _, err := tt.args.service.AddSale(tt.args.saleDTO); (err != nil) != tt.wantErr ||
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
	newSale := sale.CreateSale(saleId, "newSale", "123 road st")

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
			_ = repo.AddSale(newSale)
			got, err := tt.args.service.GetSaleById(tt.args.saleId)
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
					Address: "123 road st",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid sale empty name",
			args: args{
				saleDTO: requests.SaleRequest{
					Name:    "",
					Address: "123 road st",
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
					Address: "123 road st",
				},
			},
			wantErr:     true,
			wantErrText: "invalid sale",
		},
		{
			name: "invalid sale empty address",
			args: args{
				saleDTO: requests.SaleRequest{
					Name:    "Best sale in the east",
					Address: "",
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
