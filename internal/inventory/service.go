package inventory

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
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

func getItem(db *sql.DB, itemID string) (models.Item, error) {
	q := `select id, owner_id, title, description, image, listed from items where id = ?`
	item := models.Item{}

	row := db.QueryRow(q, itemID)
	if err := row.Scan(&item.ID, &item.OwnerID, &item.Title, &item.Description, &item.ImageURL, &item.Listed); err != nil {
		return item, err
	}

	return item, nil
}

func isOwner(db *sql.DB, itemID string, user models.User) bool {
	item, err := getItem(db, itemID)
	if err != nil {
		return false
	}
	return item.OwnerID == user.ID
}

func saveFile(f multipart.File, path string) error {

	// Basic file type sniff
	buff := make([]byte, 512)
	n, _ := f.Read(buff)
	if ct := http.DetectContentType(buff[:n]); (ct != "image/jpeg") && (ct != "image/png") {
		return errors.New("[addItem]: invalid file type")
	}

	// reset file seeker position
	if seeker, ok := f.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return err
		}
	}

	// Decode Image
	img, ftype, err := image.Decode(f)
	if err != nil {
		return err
	}

	// Create file on system
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	fmt.Println("[SYSTEM]: CREATE " + path)

	// Copy image to system file
	// _, err = io.Copy(dst, f)
	// if err != nil {
	// 	return  err
	// }

	switch ftype {
	case "png":
		// Encode Image to PNG
		encoder := png.Encoder{CompressionLevel: png.BestSpeed}
		if err = encoder.Encode(dst, img); err != nil {
			dst.Close()
			os.Remove(path)
			return err
		}
	case "jpeg":
		// Encode Image to JPEG
		if err = jpeg.Encode(dst, img, &jpeg.Options{Quality: 70}); err != nil {
			dst.Close()
			os.Remove(path)
			return err
		}
	default:
		return errors.New("Invalid Image Type")
	}

	return nil
}

func addItem(db *sql.DB, f multipart.File, item models.Item, dir string) (models.Item, error) {

	// Basic input validation
	if item.Title == "" {
		return item, errors.New("[addItem] No Item Title Provided")
	}
	if item.Description == "" {
		return item, errors.New("[addItem] No Item Description Provided")
	}

	// Generate ItemID
	id, err := generateID(16)
	if err != nil {
		return item, err
	}
	fileName := id

	path := filepath.Join(dir, fileName)

	if err = saveFile(f, path); err != nil {
		return item, err
	}

	// Create DB entry
	item.ID = id
	item.ImageURL = filepath.Join("images", fileName)

	q := `insert into items(id, owner_id, title, description, image) values (?, ?, ?, ?,?)`
	if _, err := db.Exec(q, item.ID, item.OwnerID, item.Title, item.Description, item.ImageURL); err != nil {
		return item, err
	}

	fmt.Println("[DB]: ADD ITEM(" + item.ID + ")")
	return item, nil
}

func deleteItem(db *sql.DB, itemID string, directory string) error {
	// Delete File
	file := filepath.Join(directory, itemID)
	if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		fmt.Println("[SYSTEM]: DELETE " + file)
	}

	// Delete DB Entry
	q := `delete from items where id = ?`
	if _, err := db.Exec(q, itemID); err != nil {
		return err
	}

	fmt.Println("[DB]: REMOVE ITEM(" + itemID + ")")
	return nil

}

func listItem(db *sql.DB, item models.Item) error {
	q := `update items set listed = true where id = ?`
	if _, err := db.Exec(q, item.ID); err != nil{
		return err
	}
	return nil
}

func delistItem(db *sql.DB, item models.Item) error {
	q := `update items set listed = false where id = ?`
	if _, err := db.Exec(q, item.ID); err != nil{
		return err
	}
	return nil
}
