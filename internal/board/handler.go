package board

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/encador/trady/internal/auth"
	"github.com/encador/trady/internal/bids"
	"github.com/encador/trady/internal/general"
	"github.com/encador/trady/internal/items"
	"github.com/encador/trady/internal/models"
	"github.com/encador/trady/internal/users"
	"github.com/starfederation/datastar-go/datastar"
)

type BoardSignals struct {
	SelectedItemID models.ID `json:"selectedItem"`
	ShowControls   bool      `json:"showControls"`
}

type BoardHandler struct {
	database *sql.DB
}

func NewBoardHandler(db *sql.DB) *BoardHandler {
	return &BoardHandler{database: db}
}

func (h *BoardHandler) BoardPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listings, err := items.GetAllListed(h.database)
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
		item, err := items.GetFromID(h.database, signals.SelectedItemID)
		if err != nil || !item.Listed {
			// http.Error(w, "error", http.StatusInternalServerError)
			sse.PatchElementTempl(general.MsgBox("Invalid Listing", 3), datastar.WithSelectorID("msg-box"), datastar.WithModeAppend())
			signals.SelectedItemID = ""
			signals.ShowControls = false
			sse.MarshalAndPatchSignals(signals)
			return
		}

		// Does the User own this item
		if item.IsOwner(auth.GetUser(r.Context())) {
			sse.PatchElementTempl(items.Contols(items.Data{Item: item, Details: true, Options: true, TakeBid: true, BidCount: len(bids.ForItem(item.ID))}))
		} else {
			owner, _ := users.GetUserByID(h.database, item.OwnerID)
			sse.PatchElementTempl(items.Contols(items.Data{Item: item, Details: true, Owner: owner, MakeBid: true}))
		}
		signals.ShowControls = true
		sse.MarshalAndPatchSignals(signals)
	})
}
