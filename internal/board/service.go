package board

import (
	"database/sql"

	"github.com/encador/trady/internal/models"
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

func getListing(db *sql.DB, itemID string) (models.Item, error) {
	q := `select id, owner_id, title, description, image, listed from items where id = ?`
	item := models.Item{}

	row := db.QueryRow(q, itemID)
	if err := row.Scan(&item.ID, &item.OwnerID, &item.Title, &item.Description, &item.ImageURL, &item.Listed); err != nil {
		return item, err
	}

	return item, nil
}
