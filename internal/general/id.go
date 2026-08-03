package general

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/encador/trady/internal/models"
)

func GenerateID(n int) (models.ID, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return models.ID(hex.EncodeToString(b)), nil
}
