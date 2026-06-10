package user

type UserRepository interface {
	AddUser(user User)
	GetUserByUsername(username string) (*User, error)
}
