package services

import (
	"GarageSaleAPI/domain/user"
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
					Id:       1,
					Username: "username",
					Password: "password",
					Email:    "email@email.com",
				},
			},
			wantErr: false,
		},
		{
			name: "add user with invalid email",
			args: args{
				userDTO: dto.UserDTO{
					Id:       1,
					Username: "username",
					Password: "password",
					Email:    "email",
				},
			},
			wantErr: true,
			textErr: "bad request email",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		Id:       1,
		Username: "username",
		Password: "password",
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
		textErr string
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
			textErr: "user not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := AddUser(uDTO)
			if e != nil && !tt.wantErr {
				t.Errorf("AddUser() error = %v, wantErr %v", e, tt.wantErr)
			}
			_, err := GetUserByUsername(tt.args.username)
			if (err != nil) != tt.wantErr ||
				((err != nil) && err.Error() != tt.textErr) {
				t.Errorf("GetUserByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
