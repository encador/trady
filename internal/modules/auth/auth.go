package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/encador/trady/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type authContextKey string

const ctxKey authContextKey = "key"

var secret_key []byte = []byte("super-secure-key")

// list of allowed urls without needing account
var allowList = map[string]bool{
	"/":           true,
	"/user":       true,
	"/user/new":   true,
	"/user/login": true,
}

func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := models.User{}
		if cookie, err := r.Cookie("auth"); err == nil {
			user, err = parseToken(cookie.Value)
			if err != nil {
				fmt.Println(err)
			}
		}

		url := r.URL.String()
		if user.Username == "" && !allowList[url] {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUser(ctx context.Context) string {
	user, ok := ctx.Value(ctxKey).(models.User)
	if !ok {
		return ""
	}
	return user.Username
}

func genToken(user models.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   user.Username,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret_key)
}

func parseToken(tString string) (models.User, error) {
	user := models.User{}
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tString, &claims, func(t *jwt.Token) (any, error) { return secret_key, nil })
	if err != nil {
		return user, err
	}
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		user.Username = claims.Subject
		return user, nil
	}
	return user, errors.New("[parseToken] invalid claims")
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
