package inventory

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/encador/trady/internal/models"
)

func generateID(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func getAllItems(db *sql.DB, user models.User) ([]models.Item, error) {
	q := `select id, owner_id, title, description, image, listed from items where owner_id = ?`

	items := []models.Item{}

	rows, err := db.Query(q, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.Item{}
		if err := rows.Scan(&item.ID, &item.OwnerID, &item.Title, &item.Description, &item.ImageURL, &item.Listed); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func addItem(db *sql.DB, f multipart.File, item models.Item, dir string) error {

	var fileName string
	id, err := generateID(16)
	if err != nil {
		return err
	}

	// Only allow png and jpeg
	buff := make([]byte, 512)
	n, _ := f.Read(buff)
	switch http.DetectContentType(buff[:n]) {
	case "image/jpeg":
		fileName = id + ".jpeg"
	case "image/png":
		fileName = id + ".png"
	default:
		return errors.New("[addItem]: invalid file type")
	}

	// reset file seeker position
	if seeker, ok := f.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return err
		}
	}

	// Create file on system
	path := filepath.Join(dir, fileName)
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy image to system file
	_, err = io.Copy(dst, f)
	if err != nil {
		return err
	}

	// Create DB entry
	item.ID = id
	item.ImageURL = filepath.Join("images", fileName)

	q := `insert into items(id, owner_id, title, description, image) values (?, ?, ?, ?,?)`
	if _, err := db.Exec(q, item.ID, item.OwnerID, item.Title, item.Description, item.ImageURL); err != nil {
		return err
	}

	return nil
}
