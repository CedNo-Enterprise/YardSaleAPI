package services

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/interfaces/dto"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestSaleService_AddSale(t *testing.T) {
	s := server.NewAppServer()

	type args struct {
		service *SaleService
		saleDTO dto.SaleDTO
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
				saleDTO: dto.SaleDTO{
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
				saleDTO: dto.SaleDTO{
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
		saleDTO dto.SaleDTO
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
				saleDTO: dto.SaleDTO{
					Name:    "Best sale in the east",
					Address: "123 road st",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid sale",
			args: args{
				saleDTO: dto.SaleDTO{
					Name:    "",
					Address: "123 road st",
				},
			},
			wantErr:     true,
			wantErrText: "sale name is invalid",
		},
		{
			name: "invalid sale",
			args: args{
				saleDTO: dto.SaleDTO{
					Name:    "Best sale in the east",
					Address: "",
				},
			},
			wantErr:     true,
			wantErrText: "sale address is invalid",
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

func Test_validateSaleAddress(t *testing.T) {
	type args struct {
		saleAddress string
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrText string
	}{
		{
			name: "valid sale address",
			args: args{
				saleAddress: "123 road st",
			},
			wantErr: false,
		},
		{
			name: "invalid sale address",
			args: args{
				saleAddress: "",
			},
			wantErr:     true,
			wantErrText: "sale address is empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSaleAddress(tt.args.saleAddress); (err != nil) != tt.wantErr {
				t.Errorf("validateSaleAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateSaleName(t *testing.T) {
	type args struct {
		saleName string
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrText string
	}{
		{
			name: "valid sale name",
			args: args{
				saleName: "Best sale in the east",
			},
			wantErr: false,
		},
		{
			name: "empty sale name",
			args: args{
				saleName: "",
			},
			wantErr:     true,
			wantErrText: "sale name is empty",
		},
		{
			name: "short sale name",
			args: args{
				saleName: "sale",
			},
			wantErr:     true,
			wantErrText: "sale name is too short",
		},
		{
			name: "long sale name",
			args: args{
				saleName: "This sale name is way too long for our liking and will be deemed invalid",
			},
			wantErr:     true,
			wantErrText: "sale name is too long",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSaleName(tt.args.saleName); (err != nil) != tt.wantErr ||
				(err != nil) && err.Error() != tt.wantErrText {
				t.Errorf("validateSaleName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
