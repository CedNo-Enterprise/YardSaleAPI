package memory

import (
	"GarageSaleAPI/domain/user"
	"reflect"
	"testing"
	"time"
)

func TestInMemoryUserRepository_AddUser(t *testing.T) {
	type fields struct {
		UserList []user.User
	}
	type args struct {
		user user.User
	}

	validUser := user.User{
		Username:  "username",
		Password:  "password",
		Email:     "email@email.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		textErr string
	}{
		{
			name: "add user",
			fields: fields{
				UserList: []user.User{},
			},
			args: args{
				user: validUser,
			},
			wantErr: false,
			textErr: "",
		},
		{
			name: "add duplicate user",
			fields: fields{
				UserList: []user.User{validUser},
			},
			args: args{
				user: validUser,
			},
			wantErr: true,
			textErr: "user already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := InMemoryUserRepository{
				UserList: tt.fields.UserList,
			}
			err := repo.AddUser(tt.args.user)
			if err != nil && !tt.wantErr ||
				((err != nil) && err.Error() != tt.textErr) {
				t.Errorf("InMemoryUserRepository.AddUser() error = %v, wantErr %v\ntext = %v, textErr = %v",
					err, tt.wantErr, err.Error(), tt.textErr)
			}
		})
	}
}

func TestInMemoryUserRepository_GetUserByUsername(t *testing.T) {
	type fields struct {
		UserList []user.User
	}
	type args struct {
		username string
	}

	validUser := user.User{
		Username:  "username",
		Password:  "password",
		Email:     "email@email.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *user.User
		wantErr bool
		textErr string
	}{
		{
			name: "get user by username",
			fields: fields{
				UserList: []user.User{validUser},
			},
			args: args{
				username: "username",
			},
			want:    &validUser,
			wantErr: false,
			textErr: "",
		},
		{
			name: "get user with empty list",
			fields: fields{
				UserList: []user.User{},
			},
			args: args{
				username: "username",
			},
			want:    nil,
			wantErr: true,
			textErr: "user not found",
		},
		{
			name: "get nonexistent user",
			fields: fields{
				UserList: []user.User{validUser},
			},
			args: args{
				username: "InvalidUsername",
			},
			want:    nil,
			wantErr: true,
			textErr: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := InMemoryUserRepository{
				UserList: tt.fields.UserList,
			}
			got, err := repo.GetUserByUsername(tt.args.username)
			if err != nil && !tt.wantErr ||
				((err != nil) && err.Error() != tt.textErr) {
				t.Errorf("InMemoryUserRepository.AddUser() error = %v, wantErr %v\ntext = %v, textErr = %v",
					err, tt.wantErr, err.Error(), tt.textErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserByUsername() got = %v, want %v", got, tt.want)
			}
		})
	}
}
