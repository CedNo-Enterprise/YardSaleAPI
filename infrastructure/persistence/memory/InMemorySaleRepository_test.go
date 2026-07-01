package memory

import (
	"GarageSaleAPI/domain/address"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/test"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

var validAddress = address.CreateAddress(
	"northern", nil,
	"Washington", "WS",
	"U1A 2C5", "US",
)

func TestInMemorySaleRepository_AddSale(t *testing.T) {
	type fields struct {
		SaleList []sale.Sale
	}
	type args struct {
		sale sale.Sale
		ctx  context.Context
	}

	validSale := sale.CreateSale(
		"123e4567-e89b-12d3-a456-426614174000", uuid.NewString(), "newSale",
		validAddress, time.Now(), "", time.Now(),
	)

	tests := []struct {
		name        string
		fields      fields
		args        args
		wantLength  int
		wantErr     bool
		wantErrText string
	}{
		{
			name: "add sale",
			fields: fields{
				[]sale.Sale{},
			},
			args: args{
				sale: validSale,
				ctx:  test.CreateTestContext(t),
			},
			wantLength: 1,
			wantErr:    false,
		},
		{
			name: "add duplicate sale",
			fields: fields{
				[]sale.Sale{validSale},
			},
			args: args{
				sale: validSale,
				ctx:  test.CreateTestContext(t),
			},
			wantLength:  1,
			wantErr:     true,
			wantErrText: "sale already exists",
		},
		{
			name: "add sale with timed out context",
			fields: fields{
				[]sale.Sale{},
			},
			args: args{
				sale: validSale,
				ctx:  test.CreateTimedOutTestContext(t),
			},
			wantLength:  0,
			wantErr:     true,
			wantErrText: context.DeadlineExceeded.Error(),
		},
		{
			name: "add sale with cancelled context",
			fields: fields{
				[]sale.Sale{},
			},
			args: args{
				sale: validSale,
				ctx:  test.CreateCancelledTestContext(),
			},
			wantLength:  0,
			wantErr:     true,
			wantErrText: context.Canceled.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemorySaleRepository{
				saleList: tt.fields.SaleList,
			}

			err := repo.Save(tt.args.ctx, tt.args.sale)

			if (err != nil) != tt.wantErr ||
				((err != nil) && err.Error() != tt.wantErrText) {
				t.Errorf("AddSale() error = %v, wantErr %v\ntext = %v, textErr = %v",
					err, tt.wantErr, err.Error(), tt.wantErrText)
			}

			if len(repo.saleList) != tt.wantLength {
				t.Errorf("len(sellerList) = %d, want %d", len(repo.saleList), tt.wantLength)
			}
		})
	}
}

func TestInMemorySaleRepository_GetSaleById(t *testing.T) {
	type fields struct {
		SaleList []sale.Sale
	}
	type args struct {
		id  string
		ctx context.Context
	}

	validId := "123e4567-e89b-12d3-a456-426614174000"
	validSale := sale.CreateSale(
		validId, uuid.NewString(), "newSale",
		validAddress, time.Now(), "", time.Now(),
	)

	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *sale.Sale
		wantErr     bool
		wantErrText string
	}{
		{
			name: "get existing sale",
			fields: fields{
				[]sale.Sale{validSale},
			},
			args: args{
				id:  validId,
				ctx: test.CreateTestContext(t),
			},
			want:    &validSale,
			wantErr: false,
		},
		{
			name: "get nonexistent sale",
			fields: fields{
				[]sale.Sale{
					sale.CreateSale(
						uuid.NewString(), uuid.NewString(), "different sale",
						validAddress, time.Now(), "", time.Now(),
					),
				},
			},
			args: args{
				id:  validId,
				ctx: test.CreateTestContext(t),
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "sale not found",
		},
		{
			name: "get sale with empty list",
			fields: fields{
				[]sale.Sale{},
			},
			args: args{
				id:  validId,
				ctx: test.CreateTestContext(t),
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "sale not found",
		},
		{
			name: "get sale with timed out context",
			fields: fields{
				[]sale.Sale{validSale},
			},
			args: args{
				id:  validId,
				ctx: test.CreateTimedOutTestContext(t),
			},
			want:        nil,
			wantErr:     true,
			wantErrText: context.DeadlineExceeded.Error(),
		},
		{
			name: "get sale with cancelled context",
			fields: fields{
				[]sale.Sale{validSale},
			},
			args: args{
				id:  validId,
				ctx: test.CreateCancelledTestContext(),
			},
			want:        nil,
			wantErr:     true,
			wantErrText: context.Canceled.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemorySaleRepository{
				saleList: tt.fields.SaleList,
			}

			got, err := repo.GetById(tt.args.ctx, tt.args.id)

			if (err != nil) != tt.wantErr ||
				((err != nil) && err.Error() != tt.wantErrText) {
				t.Errorf("GetSaleById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSaleById() got = %v, want %v", got, tt.want)
			}
		})
	}
}
