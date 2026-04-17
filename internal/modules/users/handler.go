// HandleUserPage

// HandleAdd
// HandleRemove
// HandleLogin
// HandleLogout

package users

import (
	"database/sql"
	"fmt"
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
		// layout.Base(SignupForm()).Render(r.Context(), w)
		layout.Base(userPage()).Render(r.Context(), w)
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
		msgs, err := addUser(user, h.database)
		if err != nil {
			fmt.Println(err)
			component.MsgBox(msgs, 3).Render(r.Context(), w)
		} else {
			component.MsgBox(msgs, 1).Render(r.Context(), w)
			fmt.Println("[HandleAdd]: New User Added")
		}
	})
}
