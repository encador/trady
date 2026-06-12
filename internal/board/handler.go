package board

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/encador/trady/internal/general"
	"github.com/starfederation/datastar-go/datastar"
)

type BoardSignals struct {
	SelectedListingID string `json:"selectedListing"`
	ShowControls      bool   `json:"showControls"`
}

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
			Content:  BoardPage(listings),
			Contorls: ListingControl(),
			URL:      "/board",
		}
		general.Base(opt).Render(r.Context(), w)
	})
}

func (h *BoardHandler) HandleSelect() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "only POST", http.StatusMethodNotAllowed)
			return
		}
		signals := &BoardSignals{}
		datastar.ReadSignals(r, signals)

		sse := datastar.NewSSE(w, r)
		l, err := getListing(h.database, signals.SelectedListingID)
		if err != nil || !l.Listed {
			// http.Error(w, "error", http.StatusInternalServerError)
			sse.PatchElementTempl(general.MsgBox("Invalid Listing", 3), datastar.WithSelectorID("msg-box"), datastar.WithModeAppend())
			signals.SelectedListingID=""
			signals.ShowControls=false
			sse.MarshalAndPatchSignals(signals)
			return
		}

		sse.PatchElementTempl(ListingContols(l))
		signals.ShowControls = true
		sse.MarshalAndPatchSignals(signals)
	})
}
