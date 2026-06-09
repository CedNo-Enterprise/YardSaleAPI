package memory

import (
	"GarageSaleAPI/domain/user"
	"reflect"
	"testing"
)

func TestInMemoryUserRepository_AddUser(t *testing.T) {
	type fields struct {
		UserList []user.User
	}
	type args struct {
		user user.User
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := InMemoryUserRepository{
				UserList: tt.fields.UserList,
			}
			repo.AddUser(tt.args.user)
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
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *user.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := InMemoryUserRepository{
				UserList: tt.fields.UserList,
			}
			got, err := repo.GetUserByUsername(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserByUsername() got = %v, want %v", got, tt.want)
			}
		})
	}
}
