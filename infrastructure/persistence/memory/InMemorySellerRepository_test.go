package memory

import (
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/test"
	"context"
	"reflect"
	"testing"
	"time"
)

var validSeller = seller.CreateSeller("1", "1", time.Now())

func TestInMemorySellerRepository_Save(t *testing.T) {
	type fields struct {
		sellerList []seller.Seller
	}
	type args struct {
		ctx    context.Context
		seller seller.Seller
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantLength  int
		wantErr     bool
		wantErrText string
	}{
		{
			name: "save seller",
			fields: fields{
				sellerList: []seller.Seller{},
			},
			args: args{
				ctx:    test.CreateTestContext(t),
				seller: validSeller,
			},
			wantLength: 1,
			wantErr:    false,
		},
		{
			name: "save duplicate seller",
			fields: fields{
				sellerList: []seller.Seller{validSeller},
			},
			args: args{
				ctx:    test.CreateTestContext(t),
				seller: validSeller,
			},
			wantLength:  1,
			wantErr:     true,
			wantErrText: "seller already exists",
		},
		{
			name: "save seller with cancelled context",
			fields: fields{
				sellerList: []seller.Seller{},
			},
			args: args{
				ctx:    test.CreateCancelledTestContext(),
				seller: validSeller,
			},
			wantLength:  0,
			wantErr:     true,
			wantErrText: context.Canceled.Error(),
		},
		{
			name: "save seller with expired context",
			fields: fields{
				sellerList: []seller.Seller{},
			},
			args: args{
				ctx:    test.CreateTimedOutTestContext(t),
				seller: validSeller,
			},
			wantLength:  0,
			wantErr:     true,
			wantErrText: context.DeadlineExceeded.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemorySellerRepository{
				sellerList: tt.fields.sellerList,
			}

			err := repo.Save(tt.args.ctx, tt.args.seller)
			if (err != nil) != tt.wantErr ||
				((err != nil) && err.Error() != tt.wantErrText) {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(repo.sellerList) != tt.wantLength {
				t.Errorf("len(sellerList) = %d, want %d", len(repo.sellerList), tt.wantLength)
			}
		})
	}
}

func TestInMemorySellerRepository_GetById(t *testing.T) {
	type fields struct {
		sellerList []seller.Seller
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *seller.Seller
		wantErr     bool
		wantErrText string
	}{
		{
			name: "get existing seller",
			fields: fields{
				sellerList: []seller.Seller{validSeller},
			},
			args: args{
				ctx: test.CreateTestContext(t),
				id:  "1",
			},
			want:    &validSeller,
			wantErr: false,
		},
		{
			name: "get nonexisting seller",
			fields: fields{
				sellerList: []seller.Seller{},
			},
			args: args{
				ctx: test.CreateTestContext(t),
				id:  "1",
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "seller not found",
		},
		{
			name: "get seller with cancelled context",
			fields: fields{
				sellerList: []seller.Seller{},
			},
			args: args{
				ctx: test.CreateCancelledTestContext(),
				id:  "1",
			},
			want:        nil,
			wantErr:     true,
			wantErrText: context.Canceled.Error(),
		},
		{
			name: "get seller with expired context",
			fields: fields{
				sellerList: []seller.Seller{},
			},
			args: args{
				ctx: test.CreateTimedOutTestContext(t),
				id:  "1",
			},
			want:        nil,
			wantErr:     true,
			wantErrText: context.DeadlineExceeded.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemorySellerRepository{
				sellerList: tt.fields.sellerList,
			}
			got, err := repo.GetById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr ||
				((err != nil) && err.Error() != tt.wantErrText) {
				t.Errorf("GetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetById() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemorySellerRepository_GetByUserId(t *testing.T) {
	type fields struct {
		sellerList []seller.Seller
	}
	type args struct {
		ctx    context.Context
		userId string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *seller.Seller
		wantErr     bool
		wantErrText string
	}{
		{
			name: "get existing seller",
			fields: fields{
				sellerList: []seller.Seller{validSeller},
			},
			args: args{
				ctx:    test.CreateTestContext(t),
				userId: "1",
			},
			want:    &validSeller,
			wantErr: false,
		},
		{
			name: "get nonexisting seller",
			fields: fields{
				sellerList: []seller.Seller{},
			},
			args: args{
				ctx:    test.CreateTestContext(t),
				userId: "1",
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "seller not found",
		},
		{
			name: "get seller with expired context",
			fields: fields{
				sellerList: []seller.Seller{},
			},
			args: args{
				ctx:    test.CreateTimedOutTestContext(t),
				userId: "1",
			},
			want:        nil,
			wantErr:     true,
			wantErrText: context.DeadlineExceeded.Error(),
		},
		{
			name: "get seller with cancelled context",
			fields: fields{
				sellerList: []seller.Seller{},
			},
			args: args{
				ctx:    test.CreateCancelledTestContext(),
				userId: "1",
			},
			want:        nil,
			wantErr:     true,
			wantErrText: context.Canceled.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemorySellerRepository{
				sellerList: tt.fields.sellerList,
			}
			got, err := repo.GetByUserId(tt.args.ctx, tt.args.userId)
			if (err != nil) != tt.wantErr ||
				((err != nil) && err.Error() != tt.wantErrText) {
				t.Errorf("GetByUserId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByUserId() got = %v, want %v", got, tt.want)
			}
		})
	}
}
