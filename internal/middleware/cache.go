package middleware

import (
	"net/http"
	"strconv"
	"time"
)

// Add cache headers
func Cache1(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" || r.Method == "HEAD" {
			age := strconv.FormatInt(int64((1 * time.Hour).Seconds()), 10)
			w.Header().Set("Cache-Control", "public, max-age="+age)
		}
		h.ServeHTTP(w, r)
	})
}
