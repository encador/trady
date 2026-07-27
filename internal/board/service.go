package board

import (
	"database/sql"

	"github.com/encador/trady/internal/inventory"
	"github.com/encador/trady/internal/models"
	"github.com/encador/trady/internal/users"
)

func getAllListings(db *sql.DB) ([]models.Item, error) {
	listings := []models.Item{}
	q := `select id, owner_id, title, description, image, listed from items where listed = true`

	rows, err := db.Query(q)
	if err != nil {
		return listings, err
	}

	for rows.Next() {
		l := models.Item{}
		if err := rows.Scan(&l.ID, &l.OwnerID, &l.Title, &l.Description, &l.ImageURL, &l.Listed); err != nil {
			return listings, err
		}
		listings = append(listings, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return listings, nil
}

func getListing(db *sql.DB, itemID string) (models.Listing, error) {
	listing := models.Listing{}
	item, err := inventory.GetItem(db, itemID)
	if err != nil {
		return listing, err
	}
	user, err := users.GetUserByID(db, item.OwnerID)
	if err != nil {
		return listing, err
	}

	listing.User = user
	listing.Item = item
	return listing, nil
}
