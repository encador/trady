package bids

import (
	"database/sql"

	"github.com/encador/trady/internal/models"
)

// All bids for specific Item ID
func ForItem(db *sql.DB, itemID models.ID) []models.Item {
	return []models.Item{}
}

// All bids made by User ID
func ByUser(db *sql.DB, userID models.ID) []models.Item {
	return []models.Item{}
}

// All bids made by User ID for Item ID
func ByUserForItem(db *sql.DB, userID models.ID, itemID models.ID) []models.Item {
	return []models.Item{}
}

func PlaceBid(db *sql.DB, userID models.ID, itemID models.ID, targetID models.ID) error {
	return nil
}
