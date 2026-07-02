package services

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/domain/seller"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/test"
	"context"
	"reflect"
	"testing"
	"time"
)

func TestSellerService_AddSeller(t *testing.T) {
	s := server.NewAppServer()
	userRepo := *s.GetUserRepository()
	_ = userRepo.Save(
		test.CreateTestContext(t),
		user.CreateUser("user", "password", "email@email.com", time.Now()),
	)

	type fields struct {
		sellerRepository seller.SellerRepository
		userRepository   user.UserRepository
	}
	type args struct {
		ctx    context.Context
		userId string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantErrText string
	}{
		{
			name: "Add seller",
			fields: fields{
				sellerRepository: *s.GetSellerRepository(),
				userRepository:   userRepo,
			},
			args: args{
				ctx:    test.CreateTestContext(t),
				userId: "user",
			},
			wantErr: false,
		},
		{
			name: "Add seller with invalid userId",
			fields: fields{
				sellerRepository: *s.GetSellerRepository(),
				userRepository:   userRepo,
			},
			args: args{
				ctx:    test.CreateTestContext(t),
				userId: "invalidUserId",
			},
			wantErr:     true,
			wantErrText: "error adding seller",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &SellerService{
				sellerRepository: tt.fields.sellerRepository,
				userRepository:   tt.fields.userRepository,
			}
			got, err := service.AddSeller(tt.args.ctx, tt.args.userId)
			if (err != nil) != tt.wantErr || (err != nil) && err.Error() != tt.wantErrText {
				t.Errorf("AddSeller() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil && got == nil {
				t.Errorf("AddSeller() got = %v, want userId", got)
			}
		})
	}
}

func TestSellerService_GetSellerById(t *testing.T) {
	s := server.NewAppServer()
	sellerRepo := *s.GetSellerRepository()
	newSeller := seller.CreateSeller("1", "1", time.Now())

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
			want:    &newSeller,
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
			_ = sellerRepo.Save(tt.args.ctx, newSeller)
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
	s := server.NewAppServer()
	userRepo := *s.GetUserRepository()
	_ = userRepo.Save(
		test.CreateTestContext(t),
		user.CreateUser("user", "password", "email@email.com", time.Now()),
	)
	sellerRepo := *s.GetSellerRepository()
	newSeller := seller.CreateSeller("1", "user", time.Now())

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
				userId: "user",
			},
			want:    &newSeller,
			wantErr: false,
		},
		{
			name: "Get seller with invalid userId",
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
			_ = sellerRepo.Save(tt.args.ctx, newSeller)
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
