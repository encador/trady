// HandleUserPage

// HandleAdd
// HandleRemove
// HandleLogin
// HandleLogout

package users

import (
	"database/sql"
	"net/http"

	"github.com/encador/trady/internal/models"
	"github.com/encador/trady/internal/templ/component"
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

func (h *UserHandler) HandleAdd() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST"{
			http.NotFoundHandler().ServeHTTP(w,r)
			return
		}
		r.ParseForm()
		form := r.PostForm
		user := models.User{
			Username: form.Get("username"),
			Password: form.Get("password"),
		}
		errors := addUser(user, h.database)
		component.MsgBox(errors, 3).Render(r.Context(), w)
	})
}
