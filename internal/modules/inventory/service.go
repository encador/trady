package inventory

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
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

func addItem(f multipart.File, item models.Item, dir string) error {

	var fileName string
	id, err := generateID(16)
	if err != nil {
		return err
	}
	item.ID = id

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

	fmt.Println(fileName)
	// Create file on system
	dst, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		return err
	}
	defer dst.Close()

	// reset file seeker position
	if seeker, ok := f.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return err
		}
	}

	// Copy image to system file
	_, err = io.Copy(dst, f)
	return err

}
