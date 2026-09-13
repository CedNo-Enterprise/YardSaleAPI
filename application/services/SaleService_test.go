package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/infrastructure/persistence/memory"
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
	"northern", nil,
	"Washington", "WS", "U1A 2C5", "US",
)

const (
	saleOwnerUserId = "cccccccc-0000-4000-8000-000000000001"
	saleOwnerId     = "cccccccc-0000-4000-8000-000000000002"
	otherUserId     = "cccccccc-0000-4000-8000-000000000003"
	absentSellerId  = "cccccccc-0000-4000-8000-00000000dead"
	saleOwnerName   = "seedvendor"
)

func newSaleServiceWithSeller(t *testing.T) (*SaleService, *memory.InMemorySaleRepository) {
	t.Helper()

	saleRepo := &memory.InMemorySaleRepository{}
	sellerRepo := &memory.InMemorySellerRepository{}
	owner := seller.CreateSeller(saleOwnerId, saleOwnerUserId, saleOwnerName, time.Now())
	if err := sellerRepo.Create(test.CreateTestContext(t), owner); err != nil {
		t.Fatalf("seeding seller: %v", err)
	}

	return NewSaleService(saleRepo, sellerRepo), saleRepo
}

func ownedSaleRequest() requests.SaleRequest {
	return requests.SaleRequest{
		SellerId: saleOwnerId,
		Name:     "Best sale in the east",
		Address:  validAddressRequest,
		Date:     time.Now(),
	}
}

func TestSaleService_AddSale(t *testing.T) {
	type args struct {
		userId  string
		saleDTO requests.SaleRequest
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrKind apperror.Kind
	}{
		{
			name: "add valid sale",
			args: args{
				userId:  saleOwnerUserId,
				saleDTO: ownedSaleRequest(),
			},
			wantErr: false,
		},
		{
			name: "add invalid sale",
			args: args{
				userId: saleOwnerUserId,
				saleDTO: requests.SaleRequest{
					Name:    "",
					Address: validAddressRequest,
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name: "add sale under a seller the caller does not own",
			args: args{
				userId:  otherUserId,
				saleDTO: ownedSaleRequest(),
			},
			wantErr:     true,
			wantErrKind: apperror.KindForbidden,
		},
		{
			name: "add sale under a seller that does not exist",
			args: args{
				userId: saleOwnerUserId,
				saleDTO: requests.SaleRequest{
					SellerId: absentSellerId,
					Name:     "Best sale in the east",
					Address:  validAddressRequest,
					Date:     time.Now(),
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := test.CreateTestContext(t)
			service, _ := newSaleServiceWithSeller(t)

			_, err := service.AddSale(ctx, tt.args.userId, tt.args.saleDTO)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddSale() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				test.AssertKind(t, err, tt.wantErrKind)
			}
		})
	}
}

func TestSaleService_GetSaleById(t *testing.T) {
	repo := &memory.InMemorySaleRepository{}
	saleId := uuid.NewString()
	newSale := sale.CreateSale(
		saleId, uuid.NewString(), "newSale",
		validAddress, time.Now(), "", time.Now(),
	)

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
				service: NewSaleService(repo, &memory.InMemorySellerRepository{}),
				saleId:  saleId,
			},
			want:    newSale,
			wantErr: false,
		},
		{
			name: "Get nonexistent sale by id",
			args: args{
				service: NewSaleService(repo, &memory.InMemorySellerRepository{}),
				saleId:  "123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := test.CreateTestContext(t)
			_ = repo.Create(ctx, newSale)
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
		wantErrKind apperror.Kind
	}{
		{
			name: "valid sale",
			args: args{
				saleDTO: requests.SaleRequest{
					SellerId: uuid.NewString(),
					Name:     "Best sale in the east",
					Address:  validAddressRequest,
					Date:     time.Now(),
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
			wantErrKind: apperror.KindInvalid,
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
			wantErrKind: apperror.KindInvalid,
		},
		{
			name: "invalid sale empty address",
			args: args{
				saleDTO: requests.SaleRequest{
					Name: "Best sale in the east",
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSale(tt.args.saleDTO)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSale() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				test.AssertKind(t, err, tt.wantErrKind)
			}
		})
	}
}

func TestSaleService_malformedSaleIdIsNotFound(t *testing.T) {
	service := NewSaleService(&memory.InMemorySaleRepository{}, &memory.InMemorySellerRepository{})
	ctx := test.CreateTestContext(t)

	_, err := service.GetSaleById(ctx, "ghost")

	test.AssertKind(t, err, apperror.KindNotFound)
}

func TestSaleService_malformedSellerIdIsInvalid(t *testing.T) {
	service := NewSaleService(&memory.InMemorySaleRepository{}, &memory.InMemorySellerRepository{})
	ctx := test.CreateTestContext(t)

	_, err := service.AddSale(ctx, saleOwnerUserId, requests.SaleRequest{
		SellerId: "ghost",
		Name:     "Best sale in the east",
		Address:  validAddressRequest,
		Date:     time.Now(),
	})

	test.AssertKind(t, err, apperror.KindInvalid)
}
