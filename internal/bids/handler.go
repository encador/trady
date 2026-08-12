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

		var pickerItems []models.Item

		if item.IsOwner(user) {
			pickerItems = ForItem(h.database, item.ID)
		} else {
			pickerItems, err = items.GetAllForUser(h.database, user.ID)
			if err != nil {
				http.Error(w, "error", http.StatusInternalServerError)
				return
			}
		}

		isTaker := item.IsOwner(user)

		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(Picker(pickerItems, isTaker), datastar.WithModeReplace())
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

		// Close Picker Menu
		sse := datastar.NewSSE(w, r)
		signals.ShowPicker = false
		signals.PickerItemID = ""
		sse.MarshalAndPatchSignals(signals)

		if err := PlaceBid(h.database, user.ID, bidItem.ID, targetItem.ID); err != nil {
			fmt.Println(err.Error())
			// http.Error(w, "error placing bid", http.StatusInternalServerError)
			sse.PatchElementTempl(general.MsgBox("Bid Place Error", 3), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
			return
		}

		sse.PatchElementTempl(general.MsgBox("Bid Placed", 1), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
		sse.PatchElementTempl(items.MakeBid(bidItem))

	})
}

func (h *BidHandler) HandleRemove() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signals := BidSignals{}
		datastar.ReadSignals(r, &signals)
		user := auth.GetUser(r.Context())
		sse := datastar.NewSSE(w, r)

		targetItem, err := items.GetFromID(h.database, signals.SelectedItemID)
		if err != nil {
			sse.PatchElementTempl(general.MsgBox("Invalid Item", 3), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
			return
		}

		if err := RemoveByUserForItem(h.database, user.ID, targetItem.ID); err != nil {
			sse.PatchElementTempl(general.MsgBox("Bid Remove Error", 3), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
			return
		}

		sse.PatchElementTempl(general.MsgBox("Bid Removed", 2), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
		sse.PatchElementTempl(items.MakeBid(models.Item{}))

	})
}

func (h *BidHandler) HandleReject() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signals := BidSignals{}
		datastar.ReadSignals(r, &signals)

		sse := datastar.NewSSE(w, r)
		signals.ShowPicker = false
		sse.MarshalAndPatchSignals(signals)

		user := auth.GetUser(r.Context())
		item, err := items.GetFromID(h.database, signals.SelectedItemID)
		bid, err2 := items.GetFromID(h.database, signals.PickerItemID)
		if err != nil || err2 != nil || !item.IsOwner(user) {
			sse.PatchElementTempl(general.MsgBox("Bid Removal Error", 3), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
			return
		}

		if err := RemoveBidForItem(h.database, bid, item); err != nil {
			sse.PatchElementTempl(general.MsgBox("Bid Removal Error", 3), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
			return
		}

		sse.PatchElementTempl(items.TakeBid(len(ForItem(h.database, item.ID))), datastar.WithModeReplace())
		sse.PatchElementTempl(general.MsgBox("Removed Bid", 2), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())

	})
}

func (h *BidHandler) HandleRejectAll() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signals := BidSignals{}
		datastar.ReadSignals(r, &signals)
		sse := datastar.NewSSE(w, r)
		user := auth.GetUser(r.Context())
		item, err := items.GetFromID(h.database, signals.SelectedItemID)
		if err != nil || !item.IsOwner(user) {
			// http.Error(w, "error", http.StatusInternalServerError)
			sse.PatchElementTempl(general.MsgBox("Bid Removal Error", 3), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
			return
		}

		if err := RemoveAllForItem(h.database, item); err != nil {
			sse.PatchElementTempl(general.MsgBox("Bid Removal Error", 3), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
			return
		}

		sse.PatchElementTempl(items.TakeBid(0), datastar.WithModeReplace())
		sse.PatchElementTempl(general.MsgBox("Removed All Bids", 2), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())

	})
}
