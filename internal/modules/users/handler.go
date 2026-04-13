// HandleUserPage

// HandleAdd
// HandleRemove
// HandleLogin
// HandleLogout

package users

import (
	"database/sql"
	"net/http"

	"github.com/encador/trady/internal/templ/layout"
)

type UserHandler struct {
	database *sql.DB
}

func NewHandler(db *sql.DB) *UserHandler {
	return &UserHandler{database: db}
}

func (h *UserHandler) HandleUserPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		layout.Base().Render(r.Context(), w)
	})
}
