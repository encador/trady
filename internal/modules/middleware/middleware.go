package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/encador/trady/internal/models"
	"github.com/encador/trady/internal/modules/auth"
	"github.com/encador/trady/internal/modules/users"
)

// Map url to security level
// use -1 for guest instead of 0, so that security is explicit
// -1: guest (no user)
// 0: invalid urls
// 1: normal user
var secLevel = map[string]int{
	"/user/login":         -1,
	"/user/new":           -1,
	"/static/datastar.js": -1,

	"/user":        1,
	"/user/logout": 1,
	"/":            1,
}

// List of urls that redirect to Login when not logged-in
var validRedirect = map[string]bool{
	"/":     true,
	"/user": true,
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
		url := r.URL.String()
		t := time.Now().Format("15:04:05")
		fmt.Printf("[%s] [%s:%d] %s (%d): %s\n", t, user.Username, user.Security, r.Method, secLevel[url], url)

		// Deny request if URL not explicitly listed in secLevel
		if secLevel[url] == 0 {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}

		if user.Security < secLevel[url] {
			if !validRedirect[url] {
				http.NotFoundHandler().ServeHTTP(w, r)
				return
			}

			users.SetRedirectCookie(url, w)
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		ctx := auth.UpdateContext(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
