package board

import (
	"database/sql"
	"net/http"

	"github.com/encador/trady/internal/general"
)

type BoardHandler struct {
	database *sql.DB
}

func NewBoardHandler(db *sql.DB) *BoardHandler {
	return &BoardHandler{database: db}
}

func (h *BoardHandler) BoardPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opt := general.Options{
			Content: general.Hello(""),
			URL: "/board",
		}
		general.Base(opt).Render(r.Context(), w)
	})
}
