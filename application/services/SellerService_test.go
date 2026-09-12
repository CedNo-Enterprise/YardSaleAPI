package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/test"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSellerService_AddSeller(t *testing.T) {
	userRepo := &memory.InMemoryUserRepository{}
	userId := uuid.NewString()
	_ = userRepo.Create(
		test.CreateTestContext(t),
		user.CreateUser(userId, "user", "password", "email@email.com", time.Now()),
	)

	type fields struct {
		sellerRepository seller.SellerRepository
		userRepository   user.UserRepository
	}
	type args struct {
		ctx      context.Context
		userId   string
		username string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantErrKind apperror.Kind
	}{
		{
			name: "Add seller",
			fields: fields{
				sellerRepository: &memory.InMemorySellerRepository{},
				userRepository:   userRepo,
			},
			args: args{
				ctx:      test.CreateTestContext(t),
				userId:   userId,
				username: "user",
			},
			wantErr: false,
		},
		{
			name: "Add seller with invalid user id",
			fields: fields{
				sellerRepository: &memory.InMemorySellerRepository{},
				userRepository:   userRepo,
			},
			args: args{
				ctx:      test.CreateTestContext(t),
				userId:   uuid.NewString(),
				username: "user",
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &SellerService{
				sellerRepository: tt.fields.sellerRepository,
				userRepository:   tt.fields.userRepository,
			}
			got, err := service.AddSeller(tt.args.ctx, tt.args.userId, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddSeller() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil && got == nil {
				t.Errorf("AddSeller() got = %v, want sellerId", got)
			}

			if tt.wantErr {
				test.AssertKind(t, err, tt.wantErrKind)
			}
		})
	}
}

func TestSellerService_GetSellerById(t *testing.T) {
	sellerRepo := &memory.InMemorySellerRepository{}
	newSeller := seller.CreateSeller("1", "1", "username", time.Now())

	type args struct {
		ctx      context.Context
		sellerId string
	}
	tests := []struct {
		name    string
		args    args
		want    *seller.Seller
		wantErr bool
	}{
		{
			name: "Get added seller",
			args: args{
				ctx:      test.CreateTestContext(t),
				sellerId: "1",
			},
			want:    newSeller,
			wantErr: false,
		},
		{
			name: "Get non added seller",
			args: args{
				ctx:      test.CreateTestContext(t),
				sellerId: "2",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = sellerRepo.Create(tt.args.ctx, newSeller)
			service := &SellerService{sellerRepository: sellerRepo}

			got, err := service.GetSellerById(tt.args.ctx, tt.args.sellerId)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSellerById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSellerById() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSellerService_GetSellerByUserId(t *testing.T) {
	userRepo := &memory.InMemoryUserRepository{}
	userId := uuid.NewString()
	_ = userRepo.Create(
		test.CreateTestContext(t),
		user.CreateUser(userId, "user", "password", "email@email.com", time.Now()),
	)
	sellerRepo := &memory.InMemorySellerRepository{}
	newSeller := seller.CreateSeller("1", userId, "username", time.Now())

	type args struct {
		ctx    context.Context
		userId string
	}
	tests := []struct {
		name    string
		args    args
		want    *seller.Seller
		wantErr bool
	}{
		{
			name: "Get added seller",
			args: args{
				ctx:    test.CreateTestContext(t),
				userId: userId,
			},
			want:    newSeller,
			wantErr: false,
		},
		{
			name: "Get seller with invalid user id",
			args: args{
				ctx:    test.CreateTestContext(t),
				userId: "1",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = sellerRepo.Create(tt.args.ctx, newSeller)
			service := &SellerService{sellerRepository: sellerRepo}

			got, err := service.GetSellerByUserId(tt.args.ctx, tt.args.userId)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSellerByUserId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSellerByUserId() got = %v, want %v", got, tt.want)
			}
		})
	}
}
