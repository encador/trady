package models

type ID string

type User struct {
	ID       ID
	Security int
	Username string
	Password string
}

type Item struct {
	ID            ID
	OwnerID       ID
	Title         string
	Description   string
	ImageURL      string
	ImageLocation string
	Listed        bool
}

func (item Item) IsOwner(user User) bool {
	return item.OwnerID == user.ID
}
