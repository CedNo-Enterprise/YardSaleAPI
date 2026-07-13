package services

import (
	"GarageSaleAPI/application/server"
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/interfaces/requests"
	"GarageSaleAPI/test"
	"testing"
)

func TestAddUser(t *testing.T) {
	s := server.NewAppServer()
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
				userService: NewUserService(*s.GetUserRepository()),
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
				userService: NewUserService(*s.GetUserRepository()),
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
				userService: NewUserService(*s.GetUserRepository()),
				username:    "username",
			},
			wantErr: false,
		},
		{
			name: "get non-added user by username",
			args: args{
				userService: NewUserService(*s.GetUserRepository()),
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
