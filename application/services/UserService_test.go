package services

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/test"
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddUser(t *testing.T) {
	s := server.NewAppServer()
	tokenService := NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)

	type args struct {
		userService *UserService
		userDTO     requests.UserRequest
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrKind apperror.Kind
	}{
		{
			name: "add valid user",
			args: args{
				userService: NewUserService(*s.GetUserRepository(), tokenService),
				userDTO: requests.UserRequest{
					Username: "username",
					Password: "password1111111",
					Email:    "email@email.com",
				},
			},
			wantErr: false,
		},
		{
			name: "add user with invalid email",
			args: args{
				userService: NewUserService(*s.GetUserRepository(), tokenService),
				userDTO: requests.UserRequest{
					Username: "username",
					Password: "password1111111",
					Email:    "email",
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := test.CreateTestContext(t)
			t.Cleanup(func() {
				s = server.NewAppServer()
			})

			err := tt.args.userService.AddUser(ctx, tt.args.userDTO)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"AddUser()\nerror = %v, wantErr %v\ntext = %v, textErr = %v",
					err, tt.wantErr, err.Error(), tt.wantErrKind)
			}

			if tt.wantErr {
				test.AssertKind(t, err, tt.wantErrKind)
			}
		})
	}
}

func TestGetUserByUsername(t *testing.T) {
	s := server.NewAppServer()
	tokenService := NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour)

	uDTO := requests.UserRequest{
		Username: "username",
		Password: "password1111111",
		Email:    "email@email.com",
	}

	type args struct {
		userService *UserService
		username    string
	}
	tests := []struct {
		name    string
		args    args
		want    *user.User
		wantErr bool
	}{
		{
			name: "get added user by username",
			args: args{
				userService: NewUserService(*s.GetUserRepository(), tokenService),
				username:    "username",
			},
			wantErr: false,
		},
		{
			name: "get non-added user by username",
			args: args{
				userService: NewUserService(*s.GetUserRepository(), tokenService),
				username:    "fake-username",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := test.CreateTestContext(t)
			t.Cleanup(func() {
				s = server.NewAppServer()
			})

			e := tt.args.userService.AddUser(ctx, uDTO)
			if e != nil && !tt.wantErr {
				t.Errorf("AddUser() error = %v, wantErr %v", e, tt.wantErr)
			}
			_, err := tt.args.userService.GetUserByUsername(ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_validateUser(t *testing.T) {
	type args struct {
		userDTO requests.UserRequest
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrKind apperror.Kind
	}{
		{
			name: "valid user",
			args: args{
				userDTO: requests.UserRequest{
					Username: "username",
					Password: "password1111111",
					Email:    "email@email.com",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid empty username",
			args: args{
				userDTO: requests.UserRequest{
					Username: "",
					Password: "password1111111",
					Email:    "email@email.com",
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name: "invalid short username",
			args: args{
				userDTO: requests.UserRequest{
					Username: "12",
					Password: "password1111111",
					Email:    "email@email.com",
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name: "invalid long username",
			args: args{
				userDTO: requests.UserRequest{
					Username: "1234567890123456",
					Password: "password1111111",
					Email:    "email@email.com",
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name: "invalid characters in username",
			args: args{
				userDTO: requests.UserRequest{
					Username: "a$apr0cky",
					Password: "password1111111",
					Email:    "email@email.com",
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name: "invalid empty password",
			args: args{
				userDTO: requests.UserRequest{
					Username: "username",
					Password: "",
					Email:    "email@email.com",
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name: "invalid short password",
			args: args{
				userDTO: requests.UserRequest{
					Username: "username",
					Password: "12345",
					Email:    "email@email.com",
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
		{
			name: "invalid long password",
			args: args{
				userDTO: requests.UserRequest{
					Username: "username",
					Password: "12345678901234567890123456789012345678901234567890123456789012345",
					Email:    "email@email.com",
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUser(tt.args.userDTO)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				test.AssertKind(t, err, tt.wantErrKind)
			}
		})
	}
}

func TestUserService_hashPassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "hash valid password",
			args: args{
				password: "password",
			},
			wantErr: false,
		},
		{
			name: "hash invalid password",
			args: args{
				password: "passwordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpassworda",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hashPassword(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("hashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, tt.args.password) {
				t.Errorf("hashPassword() got = %v, wanted hashed password", got)
			}
		})
	}
}

func Test_comparePasswords(t *testing.T) {
	hp, _ := hashPassword("password")
	type args struct {
		hashedPassword string
		plainPassword  string
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrKind apperror.Kind
	}{
		{
			name: "compare valid password",
			args: args{
				hashedPassword: hp,
				plainPassword:  "password",
			},
			wantErr: false,
		},
		{
			name: "compare invalid password",
			args: args{
				hashedPassword: hp,
				plainPassword:  "NotTheRightPassword",
			},
			wantErr:     true,
			wantErrKind: apperror.KindUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := comparePasswords(tt.args.hashedPassword, tt.args.plainPassword); (err != nil) != tt.wantErr {
				t.Errorf("comparePasswords() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateLogin(t *testing.T) {
	type args struct {
		loginDTO requests.LoginRequest
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrKind apperror.Kind
	}{
		{
			name: "valid login request",
			args: args{
				loginDTO: requests.LoginRequest{
					Username: "username",
					Password: "validpassword",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid login request",
			args: args{
				loginDTO: requests.LoginRequest{
					Username: "username",
					Password: "p",
				},
			},
			wantErr:     true,
			wantErrKind: apperror.KindUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateLogin(tt.args.loginDTO); (err != nil) != tt.wantErr {
				t.Errorf("validateLogin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	s := server.NewAppServer()
	userRepo := *s.GetUserRepository()
	hashedPassword, _ := hashPassword("validPassword")
	newUser := user.CreateUser("username", hashedPassword, "email@email.com", time.Now())
	_ = userRepo.Save(context.Background(), newUser)
	type fields struct {
		userRepository user.UserRepository
		tokenGenerator TokenGenerator
	}
	type args struct {
		ctx      context.Context
		loginDTO requests.LoginRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "login successfully",
			fields: fields{
				*s.GetUserRepository(),
				NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour),
			},
			args: args{
				context.Background(),
				requests.LoginRequest{
					Username: "username",
					Password: "validPassword",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "login successfully",
			fields: fields{
				*s.GetUserRepository(),
				NewTokenService([]byte("f81d4fae-7dec-11d0-a765-00a0c91e6bf6"), 24*time.Hour),
			},
			args: args{
				context.Background(),
				requests.LoginRequest{
					Username: "username",
					Password: "invalidPassword",
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &UserService{
				userRepository: tt.fields.userRepository,
				tokenGenerator: tt.fields.tokenGenerator,
			}
			loginResult, err := service.Login(tt.args.ctx, tt.args.loginDTO)
			if !tt.wantErr(t, err, fmt.Sprintf("Login(%v, %v)", tt.args.ctx, tt.args.loginDTO)) {
				assert.NotNil(t, loginResult.User, "Login(%v, %v)", tt.args.ctx, tt.args.loginDTO)
				assert.NotNil(t, loginResult.ExpiresAt, "Login(%v, %v)", tt.args.ctx, tt.args.loginDTO)
				assert.NotNil(t, loginResult.AccessToken, "Login(%v, %v)", tt.args.ctx, tt.args.loginDTO)
				return
			}
		})
	}
}
