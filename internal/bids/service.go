package bids

import (
	"database/sql"

	"github.com/encador/trady/internal/items"
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

// Creates a bid for the target item using user's item
func PlaceBid(db *sql.DB, userID models.ID, itemID models.ID, targetID models.ID) error {
	// List Item if Unlisted
	if err := items.List(db, itemID); err != nil {
		return err
	}
	q := `insert into bids(target_item_id, bid_item_id) values (?,?)`
	_, err := db.Exec(q, targetID, itemID)
	if err != nil {
		return err
	}
	return nil
}
