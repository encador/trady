package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	// "github.com/encador/trady/internal/models"
	"github.com/encador/trady/internal/models"
	"github.com/encador/trady/internal/modules/auth"
	"github.com/encador/trady/internal/modules/users"
)

// list of allowed urls without needing account
var allowList = map[string]bool{
	// "/":           true,
	"/user":       true,
	"/user/new":   true,
	"/user/login": true,
}

// List of protected urls that can be redirected to
var validRedirect = map[string]bool{
	"/": true,
}

func AuthHandler(next http.Handler, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := ""
		if cookie, err := r.Cookie("auth"); err == nil {
			claims, err := auth.ParseToken(cookie.Value)
			if err != nil {
				// fmt.Println(err)
				fmt.Println("[Auth] Invalid JWT")
				auth.RemoveCookie(w)
			} else {
				username = claims.Subject
			}
		}

		user := models.User{}

		if username != "" {
			u, err := users.GetUser(username, db)
			if err != nil {
				// fmt.Println(err)
				fmt.Println("[Auth] Invalid User")
				auth.RemoveCookie(w)
			}
			user = u
		}

		// Basic Request Logging
		t := time.Now().Format("15:04:05")
		fmt.Printf("[%s] [%s] %s: %s\n", t, user.Username, r.Method, r.URL)

		url := r.URL.String()

		if user.Username == "" && !allowList[url] {

			if !validRedirect[url] {
				http.NotFoundHandler().ServeHTTP(w, r)
				return
			}

			users.SetRedirectCookie(url, w)
			http.Redirect(w, r, "/user", http.StatusSeeOther)
			return
		}

		ctx := auth.UpdateContext(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
