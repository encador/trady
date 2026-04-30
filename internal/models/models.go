package models

type User struct {
	ID       int
	Security int
	Username string
	Password string
}

type Item struct {
	ID          string
	OwnerID     int
	Title       string
	Description string
	ImageURL    string
	Listed      bool
}
