package memory

import (
	"GarageSaleAPI/domain/user"
	"GarageSaleAPI/test"
	"context"
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
		ctx  context.Context
	}

	validUser := user.CreateUser("username", "password", "email@email.com", time.Now())

	tests := []struct {
		name       string
		fields     fields
		args       args
		wantLength int
		wantErr    bool
		textErr    string
	}{
		{
			name: "add user",
			fields: fields{
				UserList: []user.User{},
			},
			args: args{
				user: validUser,
				ctx:  test.CreateTestContext(t),
			},
			wantLength: 1,
			wantErr:    false,
			textErr:    "",
		},
		{
			name: "add duplicate user",
			fields: fields{
				UserList: []user.User{validUser},
			},
			args: args{
				user: validUser,
				ctx:  test.CreateTestContext(t),
			},
			wantLength: 1,
			wantErr:    true,
			textErr:    "user already exists",
		},
		{
			name: "add user with timed out context",
			fields: fields{
				UserList: []user.User{},
			},
			args: args{
				user: validUser,
				ctx:  test.CreateTimedOutTestContext(t),
			},
			wantLength: 0,
			wantErr:    true,
			textErr:    context.DeadlineExceeded.Error(),
		},
		{
			name: "add user with cancelled context",
			fields: fields{
				UserList: []user.User{},
			},
			args: args{
				user: validUser,
				ctx:  test.CreateCancelledTestContext(),
			},
			wantLength: 0,
			wantErr:    true,
			textErr:    context.Canceled.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := InMemoryUserRepository{
				userList: tt.fields.UserList,
			}

			err := repo.AddUser(tt.args.ctx, tt.args.user)
			if err != nil && !tt.wantErr ||
				((err != nil) && err.Error() != tt.textErr) {
				t.Errorf("InMemoryUserRepository.AddUser() error = %v, wantErr %v\ntext = %v, textErr = %v",
					err, tt.wantErr, err.Error(), tt.textErr)
			}

			if len(repo.userList) != tt.wantLength {
				t.Errorf("len(userList) = %d, want %d", len(repo.userList), tt.wantLength)
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
		ctx      context.Context
	}

	validUser := user.CreateUser("username", "password", "email@email.com", time.Now())

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
				ctx:      test.CreateTestContext(t),
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
				ctx:      test.CreateTestContext(t),
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
				ctx:      test.CreateTestContext(t),
			},
			want:    nil,
			wantErr: true,
			textErr: "user not found",
		}, {
			name: "get user timed out context",
			fields: fields{
				UserList: []user.User{validUser},
			},
			args: args{
				username: "username",
				ctx:      test.CreateTimedOutTestContext(t),
			},
			want:    nil,
			wantErr: true,
			textErr: context.DeadlineExceeded.Error(),
		},
		{
			name: "get user with cancelled context",
			fields: fields{
				UserList: []user.User{validUser},
			},
			args: args{
				username: "username",
				ctx:      test.CreateCancelledTestContext(),
			},
			want:    nil,
			wantErr: true,
			textErr: context.Canceled.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := InMemoryUserRepository{
				userList: tt.fields.UserList,
			}
			got, err := repo.GetUserByUsername(tt.args.ctx, tt.args.username)
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
