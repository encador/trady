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
	"github.com/encador/trady/internal/modules/general"
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

		user := auth.GetUser(r.Context())

		opt := general.Options{
			Content: userPage(user),
			URL:     "/user",
		}
		general.Base(opt).Render(r.Context(), w)
	})
}

func (h *UserHandler) HandleLoginPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			opt := general.Options{
				Content: loginPage(),
				URL:     "/user/login",
			}
			general.Base(opt).Render(r.Context(), w)
			return

		} else {
			h.HandleLogin().ServeHTTP(w, r)
		}

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
			sse.PatchElementTempl(general.MsgBoxMultiple([]string{"Wrong Username or Password"}, 3), datastar.WithSelectorID("errors"), datastar.WithModeInner())
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

		sse := datastar.NewSSE(w, r)
		sse.Redirect("/user/login")
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
			sse.PatchElementTempl(general.MsgBoxMultiple(msgs, 2), datastar.WithSelectorID("errors"), datastar.WithModeInner())
		} else {

			if err := auth.SetCookie(user, w); err != nil {
				fmt.Println(err)
			}

			sse := datastar.NewSSE(w, r)
			sse.PatchElementTempl(general.MsgBoxMultiple(msgs, 1), datastar.WithSelectorID("errors"), datastar.WithModeInner())
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
