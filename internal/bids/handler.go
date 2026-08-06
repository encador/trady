package bids

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/encador/trady/internal/auth"
	"github.com/encador/trady/internal/general"
	"github.com/encador/trady/internal/items"
	"github.com/encador/trady/internal/models"
	"github.com/starfederation/datastar-go/datastar"
)

type BidSignals struct {
	SelectedItemID models.ID `json:"selectedItem"`
	ShowPicker     bool      `json:"showPicker"`
	PickerItemID   models.ID `json:"pickerItem"`
}

type BidHandler struct {
	database *sql.DB
}

func NewBidHandler(db *sql.DB) *BidHandler {
	return &BidHandler{database: db}
}

func (h *BidHandler) HandlePicker() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signals := &BidSignals{}
		datastar.ReadSignals(r, signals)

		user := auth.GetUser(r.Context())
		item, err := items.GetFromID(h.database, signals.SelectedItemID)
		if err != nil || !item.Listed {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
		items, err := items.GetAllForUser(h.database, user.ID)
		if err != nil {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		isTaker := item.IsOwner(user)

		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(Picker(items, isTaker), datastar.WithModeReplace())
	})
}

func (h *BidHandler) HandleMake() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signals := &BidSignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			http.Error(w, "signal error", http.StatusInternalServerError)
			return
		}

		user := auth.GetUser(r.Context())
		bidItem, err := items.GetFromID(h.database, signals.PickerItemID)
		if err != nil || !bidItem.IsOwner(user) {
			http.Error(w, "selected item error", http.StatusInternalServerError)
			return
		}
		targetItem, err := items.GetFromID(h.database, signals.SelectedItemID)
		// User cannot bid on their own items
		if err != nil || targetItem.IsOwner(user) || !targetItem.Listed {
			http.Error(w, "target item error", http.StatusInternalServerError)
			return
		}

		if err := PlaceBid(h.database, user.ID, bidItem.ID, targetItem.ID); err != nil {
			fmt.Println(err.Error())
			http.Error(w, "error placing bid", http.StatusInternalServerError)
			return
		}
		sse := datastar.NewSSE(w, r)
		signals.ShowPicker = false
		signals.PickerItemID = ""
		sse.MarshalAndPatchSignals(signals)
		sse.PatchElementTempl(general.MsgBox("Bid Placed", 1), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())

	})
}
