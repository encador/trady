package bids

import (
	"database/sql"
	"net/http"

	"github.com/encador/trady/internal/models"
	"github.com/encador/trady/internal/auth"
	"github.com/encador/trady/internal/items"
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

func (h *BidHandler) HandleBidPicker() http.Handler {
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
