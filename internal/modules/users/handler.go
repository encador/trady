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
	"time"

	"github.com/encador/trady/internal/models"
	"github.com/encador/trady/internal/modules/auth"
	"github.com/encador/trady/internal/templ/component"
	"github.com/encador/trady/internal/templ/layout"
	"github.com/starfederation/datastar-go/datastar"
)

type UserHandler struct {
	database *sql.DB
}

func NewHandler(db *sql.DB) *UserHandler {
	return &UserHandler{database: db}
}

func (h *UserHandler) HandleUserPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := auth.GetUsername(r.Context())

		// Not signed-in
		if username == "" {
			opt := layout.Options{
				Content: loginPage(),
				URL:     "/user",
			}
			layout.Base(opt).Render(r.Context(), w)
			return
		}

		user, err := GetUser(username, h.database)
		if err != nil {
			fmt.Println(err)
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}

		opt := layout.Options{
			Content: userPage(user),
			URL:     "/user",
		}
		layout.Base(opt).Render(r.Context(), w)
	})
}

func (h *UserHandler) HandleLogin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		r.ParseForm()
		form := r.PostForm
		user := models.User{
			Username: form.Get("username"),
			Password: form.Get("password"),
		}

		if err := verifyPass(user.Username, user.Password, h.database); err != nil {
			fmt.Println(err)
			sse := datastar.NewSSE(w, r)
			sse.PatchElementTempl(component.MsgBox([]string{"Wrong Username or Password"}, 3))
			return
		}
		if err := auth.SetCookie(user, w); err != nil {
			fmt.Println(err)
		}

		// sse.PatchElementTempl(component.MsgBox([]string{"Success"}, 1))
		if url, err := getRedirectURL(r); err == nil {
			UnsetRedirectCookie(w)
			sse := datastar.NewSSE(w, r)
			sse.Redirect(url)
			return
		}
		// http.Redirect(w, r, "/user", http.StatusSeeOther)
		sse := datastar.NewSSE(w, r)
		sse.Redirect("/user")
	})
}

func (h *UserHandler) HandleLogout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		auth.RemoveCookie(w)
		http.Redirect(w, r, "/user", http.StatusSeeOther)
	})
}

func (h *UserHandler) HandleAdd() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFoundHandler().ServeHTTP(w, r)
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
			sse := datastar.NewSSE(w, r)
			sse.PatchElementTempl(component.MsgBox(msgs, 2))
		} else {

			if err := auth.SetCookie(user, w); err != nil {
				fmt.Println(err)
			}

			sse := datastar.NewSSE(w, r)
			sse.PatchElementTempl(component.MsgBox(msgs, 1))
			fmt.Println("[HandleAdd]: New User Added")
			time.Sleep(time.Second * 1)

			if url, err := getRedirectURL(r); err == nil {
				UnsetRedirectCookie(w)
				sse.Redirect(url)
				return
			}

			sse.Redirect("/user")

		}
	})
}
