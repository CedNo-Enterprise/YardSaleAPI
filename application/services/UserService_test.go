package services

import (
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/infrastructure/persistence/memory"
	"GarageSaleAPI/interfaces/dto"
	"testing"
)

func TestAddUser(t *testing.T) {
	type args struct {
		userDTO dto.UserDTO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		textErr string
	}{
		{
			name: "add valid user",
			args: args{
				userDTO: dto.UserDTO{
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
				userDTO: dto.UserDTO{
					Username: "username",
					Password: "password1111111",
					Email:    "email",
				},
			},
			wantErr: true,
			textErr: "bad request email",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				UserRepository = new(memory.InMemoryUserRepository)
			})

			if err := AddUser(tt.args.userDTO); (err != nil) != tt.wantErr ||
				((err != nil) && err.Error() != tt.textErr) {
				t.Errorf(
					"AddUser()\nerror = %v, wantErr %v\ntext = %v, textErr = %v",
					err, tt.wantErr, err.Error(), tt.textErr)
			}
		})
	}
}

func TestGetUserByUsername(t *testing.T) {
	uDTO := dto.UserDTO{
		Username: "username",
		Password: "password1111111",
		Email:    "email@email.com",
	}

	type args struct {
		username string
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
				username: "username",
			},
			wantErr: false,
		},
		{
			name: "get non-added user by username",
			args: args{
				username: "fake-username",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				UserRepository = new(memory.InMemoryUserRepository)
			})

			e := AddUser(uDTO)
			if e != nil && !tt.wantErr {
				t.Errorf("AddUser() error = %v, wantErr %v", e, tt.wantErr)
			}
			_, err := GetUserByUsername(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_validateUser(t *testing.T) {
	type args struct {
		userDTO dto.UserDTO
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrText string
	}{
		{
			name: "valid user",
			args: args{
				userDTO: dto.UserDTO{
					Username: "username",
					Password: "password1111111",
					Email:    "email@email.com",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid username",
			args: args{
				userDTO: dto.UserDTO{
					Username: "",
					Password: "password1111111",
					Email:    "email@email.com",
				},
			},
			wantErr:     true,
			wantErrText: "username is invalid",
		},
		{
			name: "invalid password",
			args: args{
				userDTO: dto.UserDTO{
					Username: "username",
					Password: "",
					Email:    "email@email.com",
				},
			},
			wantErr:     true,
			wantErrText: "password is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUser(tt.args.userDTO)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUser() error = %v, wantErr %v", err, tt.wantErr)
			} else if err != nil && err.Error() != tt.wantErrText {
				t.Errorf("validateUser() error = %v, wantErrText %v", err, tt.wantErrText)
			}
		})
	}
}

func Test_validateUsername(t *testing.T) {
	type args struct {
		username string
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrText string
	}{
		{
			name: "valid username",
			args: args{
				"username",
			},
			wantErr: false,
		},
		{
			name: "empty username",
			args: args{
				"",
			},
			wantErr:     true,
			wantErrText: "username is empty",
		},
		{
			name: "username too short",
			args: args{
				"12",
			},
			wantErr:     true,
			wantErrText: "username is too short",
		},
		{
			name: "username too long",
			args: args{
				"1234567890123456",
			},
			wantErr:     true,
			wantErrText: "username is too long",
		},
		{
			name: "invalid characters in username",
			args: args{
				"a$apr0cky",
			},
			wantErr:     true,
			wantErrText: "username must only alphanumeric characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateUsername(tt.args.username); (err != nil) != tt.wantErr {
				t.Errorf("validateUsername() error = %v, wantErr %v", err, tt.wantErr)
			} else if err != nil && err.Error() != tt.wantErrText {
				t.Errorf("validateUsername() error = %v, wantErrText %v", err, tt.wantErrText)
			}
		})
	}
}

func Test_validatePassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrText string
	}{
		{
			name: "valid password",
			args: args{
				"123456789012",
			},
			wantErr: false,
		},
		{
			name: "empty password",
			args: args{
				"",
			},
			wantErr:     true,
			wantErrText: "password is empty",
		},
		{
			name: "password too short",
			args: args{
				"12345",
			},
			wantErr:     true,
			wantErrText: "password is too short",
		},
		{
			name: "password too long",
			args: args{
				"12345678901234567890123456789012345678901234567890123456789012345",
			},
			wantErr:     true,
			wantErrText: "password is too long",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePassword(tt.args.password); (err != nil) != tt.wantErr {
				t.Errorf("validatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
