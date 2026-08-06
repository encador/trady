package bids

import (
	"database/sql"

	"github.com/encador/trady/internal/items"
	"github.com/encador/trady/internal/models"
)

// All bids for specific Item ID
func ForItem(db *sql.DB, itemID models.ID) []models.Item {
	out := []models.Item{}
	q := "select bid_item_id from bids where target_item_id = ?"
	rows, _ := db.Query(q, itemID)
	for rows.Next() {
		var id models.ID
		rows.Scan(&id)
		item, err := items.GetFromID(db, id)
		if err != nil {
			return out
		}
		out = append(out, item)
	}
	return out
}

// All bids made by User ID
func ByUser(db *sql.DB, userID models.ID) []models.Item {
	return []models.Item{}
}

// All bids made by User ID for Item ID
func ByUserForItem(db *sql.DB, userID models.ID, itemID models.ID) models.Item {
	q := `select bid_item_id from bids where target_item_id = ? and bid_owner_id = ?`
	row := db.QueryRow(q, itemID, userID)
	var id models.ID
	row.Scan(&id)
	item, err := items.GetFromID(db, id)
	if err != nil {
		return models.Item{}
	}
	return item
}

// Creates a bid for the target item using user's item
func PlaceBid(db *sql.DB, userID models.ID, itemID models.ID, targetID models.ID) error {
	q := `insert into bids(target_item_id, bid_item_id, bid_owner_id) values (?,?,?)`
	_, err := db.Exec(q, targetID, itemID, userID)
	if err != nil {
		return err
	}

	// List Item if Unlisted
	if err := items.List(db, itemID); err != nil {
		return err
	}
	return nil
}

func RemoveByUserForItem(db *sql.DB, userID models.ID, itemID models.ID) error {
	q := "delete from bids where target_item_id = ? and bid_owner_id = ?"

	_, err := db.Exec(q, itemID, userID)

	return err
}
