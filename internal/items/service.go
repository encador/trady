package items

import (
	"database/sql"

	"github.com/encador/trady/internal/database"
	"github.com/encador/trady/internal/models"
)

func GetAllListed(db *sql.DB) ([]models.Item, error) {
	q := database.ItemColumns + "where listed=true"
	items := []models.Item{}

	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.Item{}
		if err := rows.Scan(&item.ID, &item.OwnerID, &item.Title, &item.Description, &item.ImageURL, &item.ImageLocation, &item.Listed); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil

}

func GetAllForUser(db *sql.DB, userID models.ID) ([]models.Item, error) {
	q := `select id, owner_id, title, description, image, location, listed from items where owner_id = ?`

	items := []models.Item{}

	rows, err := db.Query(q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.Item{}
		if err := rows.Scan(&item.ID, &item.OwnerID, &item.Title, &item.Description, &item.ImageURL, &item.ImageLocation, &item.Listed); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func GetFromID(db *sql.DB, itemID models.ID) (models.Item, error) {
	q := `select id, owner_id, title, description, image, location, listed from items where id = ?`
	item := models.Item{}

	row := db.QueryRow(q, itemID)
	if err := row.Scan(&item.ID, &item.OwnerID, &item.Title, &item.Description, &item.ImageURL, &item.ImageLocation, &item.Listed); err != nil {
		return item, err
	}

	return item, nil
}

func Add(db *sql.DB, item models.Item) error {
	q := `insert into items(id, owner_id, title, description, image, location, listed) values (?, ?, ?, ?, ?,?, ?)`
	if _, err := db.Exec(q, item.ID, item.OwnerID, item.Title, item.Description, item.ImageURL, item.ImageLocation, item.Listed); err != nil {
		return err
	}
	return nil
}

func Remove(db *sql.DB, itemID models.ID) error {
	q := `delete from items where id = ?`
	if _, err := db.Exec(q, itemID); err != nil {
		return err
	}
	return nil
}

func List(db *sql.DB, itemID models.ID) error {
	q := `update items set listed = true where id = ?`
	if _, err := db.Exec(q, itemID); err != nil {
		return err
	}
	return nil
}

func Delist(db *sql.DB, itemID models.ID) error {
	q := `update items set listed = false where id = ?`
	if _, err := db.Exec(q, itemID); err != nil {
		return err
	}
	return nil
}
