package board

import (
	"database/sql"
	"fmt"
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
		listings, err := getAllListings(h.database)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		opt := general.Options{
			Content: Board(listings),
			URL: "/board",
		}
		general.Base(opt).Render(r.Context(), w)
	})
}
