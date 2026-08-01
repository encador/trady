package board

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/encador/trady/internal/auth"
	"github.com/encador/trady/internal/bids"
	"github.com/encador/trady/internal/general"
	"github.com/encador/trady/internal/items"
	"github.com/encador/trady/internal/models"
	"github.com/starfederation/datastar-go/datastar"
)

type BoardSignals struct {
	SelectedItemID string `json:"selectedItem"`
	ShowControls   bool   `json:"showControls"`
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
			Contorls: ControlBox(),
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
		l, err := getListing(h.database, signals.SelectedItemID)
		if err != nil || !l.Item.Listed {
			// http.Error(w, "error", http.StatusInternalServerError)
			sse.PatchElementTempl(general.MsgBox("Invalid Listing", 3), datastar.WithSelectorID("msg-box"), datastar.WithModeAppend())
			signals.SelectedItemID = ""
			signals.ShowControls = false
			sse.MarshalAndPatchSignals(signals)
			return
		}

		// Does the User own this listing
		if auth.GetUser(r.Context()).ID == l.Item.OwnerID {
			sse.PatchElementTempl(items.Contols(items.Data{Item: l.Item, Details: true, Options: true, TakeBid: true, BidCount: bids.ForItem(l.Item.ID).Count}))
		} else {
			sse.PatchElementTempl(items.Contols(items.Data{Item: l.Item, Details: true, Owner: l.Owner, MakeBid: true}))
		}
		signals.ShowControls = true
		sse.MarshalAndPatchSignals(signals)
	})
}

func (h *BoardHandler) HandleBidPicker() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(bids.Picker(models.Bids{}))
		time.Sleep(time.Second)
	})
}
