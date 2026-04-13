// HandleUserPage

// HandleAdd
// HandleRemove
// HandleLogin
// HandleLogout

package users

import (
	"database/sql"
	"net/http"

	"github.com/encador/trady/internal/templ/view"
)

type UserHandler struct {
	database *sql.DB
}

func NewHandler(db *sql.DB) *UserHandler {
	return &UserHandler{database: db}
}

func (h *UserHandler) HandleUserPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view.Base().Render(r.Context(), w)
	})
}
