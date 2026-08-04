package bids

import (
	"database/sql"

	"github.com/encador/trady/internal/models"
)

// All bids for specific Item ID
func ForItem(itemID models.ID) []models.Item {
	return []models.Item{}
}

// All bids made by User ID
func ByUser(userID models.ID) []models.Item {
	return []models.Item{}
}

// All bids made by User ID for Item ID
func ByUserForItem(userID models.ID, itemID models.ID) []models.Item {
	return []models.Item{}
}

func UserItems(db *sql.DB, userID models.ID) []models.Item {

	return []models.Item{}

}
