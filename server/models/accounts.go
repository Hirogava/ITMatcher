package models

type Account struct {
	HR   *HR
	User *User
}

type HR struct {
	ID       int
	Username string
	Email    string
}

type User struct {
	ID    int
	Email string
}
