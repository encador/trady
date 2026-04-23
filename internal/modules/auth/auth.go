package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/encador/trady/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type authContextKey string

const ctxKey authContextKey = "key"

var secret_key []byte = []byte("super-secure-key")

func UpdateContext(ctx context.Context, user models.User) context.Context {
	return context.WithValue(ctx, ctxKey, user)
}

func GetUsername(ctx context.Context) string {
	user, ok := ctx.Value(ctxKey).(models.User)
	if !ok {
		return ""
	}
	return user.Username
}

func GetUser(ctx context.Context) models.User {
	user, ok := ctx.Value(ctxKey).(models.User)
	if ok {
		return user
	}
	return models.User{}
}

func genToken(user models.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   user.Username,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret_key)
}

func ParseToken(tString string) (*jwt.RegisteredClaims, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tString, &claims, func(t *jwt.Token) (any, error) { return secret_key, nil })
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("[parseToken] invalid claims")
}

func SetCookie(user models.User, w http.ResponseWriter) error {
	token, err := genToken(user)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 1),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}

func RemoveCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    "",
		Path:     "/",
		Expires:  time.Now(),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
