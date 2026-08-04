package inventory

import (
	"database/sql"
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

	"github.com/encador/trady/internal/general"
	"github.com/encador/trady/internal/items"
	"github.com/encador/trady/internal/models"
)

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
	id, err := general.GenerateID(16)
	if err != nil {
		return item, err
	}
	fileName := string(id)

	item.ID = id
	item.ImageURL = filepath.Join("images", fileName)
	item.ImageLocation = filepath.Join(dir, fileName)

	if err = saveFile(f, item.ImageLocation); err != nil {
		return item, err
	}

	// Create DB entry
	if err := items.Add(db, item); err != nil {
		return models.Item{}, err
	}

	fmt.Println("[DB]: ADD ITEM(" + item.ID + ")")
	return item, nil
}

func deleteItem(db *sql.DB, item models.Item) error {
	// Delete File
	if err := os.Remove(item.ImageLocation); err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		fmt.Println("[SYSTEM]: DELETE " + item.ImageLocation)
	}

	// Delete DB Entry
	items.Remove(db, item.ID)

	fmt.Println("[DB]: REMOVE ITEM(" + string(item.ID) + ")")
	return nil

}
