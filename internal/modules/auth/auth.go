package auth

import (
	"github.com/encador/trady/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

var secret_key []byte = []byte("super-secure-key")

func getToken(user models.User) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"Username": user.Username}).SignedString(secret_key)
}

func SetCookie(user models.User, w http.ResponseWriter) error {
	token, err := getToken(user)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    token,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}
