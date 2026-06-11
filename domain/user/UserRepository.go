package user

type UserRepository interface {
	AddUser(user User) error
	GetUserByUsername(username string) (*User, error)
}
