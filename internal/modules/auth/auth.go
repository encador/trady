package auth

import (
	"net/http"
	"time"

	"github.com/encador/trady/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

var secret_key []byte = []byte("super-secure-key")

func getToken(user models.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject: user.Username,
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret_key)
}

func SetCookie(user models.User, w http.ResponseWriter) error {
	token, err := getToken(user)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    token,
		Path: "/",
		Expires: time.Now().Add(time.Hour * 1),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}
