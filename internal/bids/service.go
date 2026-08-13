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

func RemoveBidForItem(db *sql.DB, bid, item models.Item) error {
	q := "delete from bids where target_item_id = ? and bid_item_id = ?"
	_, err := db.Exec(q, item.ID, bid.ID)
	return err
}

func RemoveAllForItem(db *sql.DB, item models.Item) error {
	q := "delete from bids where target_item_id = ?"
	_, err := db.Exec(q, item.ID)
	return err
}

// SwapItemOwnership swaps the owner_id between two users and their items
//
// Potential race condition
func SwapItemOwnership(db *sql.DB, item1, item2 models.Item, user1, user2 models.User) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	q := "update items set owner_id = ? where id = ? and owner_id = ?"
	if _, err := tx.Exec(q, user2.ID, item1.ID, user1.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(q, user1.ID, item2.ID, user2.ID); err != nil {
		return err
	}
	q = "update items set listed = false where id = ? or id = ?"
	if _, err := tx.Exec(q, item1.ID, item2.ID); err != nil {
		return err
	}

	return tx.Commit()
}
