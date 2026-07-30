package bids

import "github.com/encador/trady/internal/models"

// All bids for specific Item ID
func ForItem(id string) models.Bids {
	return models.Bids{}
}

// All bids made by User ID
func ByUser(id int) models.Bids {
	return models.Bids{}
}

// All bids made by User ID for Item ID
func ByUserForItem(userID int, itemID string) models.Bids {
	return models.Bids{}
}
