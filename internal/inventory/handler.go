package inventory

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/encador/trady/internal/auth"
	"github.com/encador/trady/internal/general"
	"github.com/encador/trady/internal/models"
	"github.com/starfederation/datastar-go/datastar"
)

type InventorySignals struct {
	SelectedItemID string `json:"selectedItem"`
	ShowControls   bool   `json:"showControls"`
}

type InventoryHandler struct {
	database  *sql.DB
	uploadDir string
}

func NewHandler(db *sql.DB, uploadDir string) (*InventoryHandler, error) {
	if info, err := os.Stat(uploadDir); err != nil || !info.IsDir() {
		return nil, errors.New("[NewHandler]: Invalid uploadDir path")
	}

	return &InventoryHandler{
		database:  db,
		uploadDir: uploadDir,
	}, nil
}

func (h *InventoryHandler) InventoryPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items, _ := getAllItems(h.database, auth.GetUser(r.Context()))
		opts := general.Options{
			Content:  InventoryPage(items),
			Contorls: InventoryControl(),
			URL:      "/inventory",
		}
		general.Base(opts).Render(r.Context(), w)
	})
}

const (
	// 5 MB image limit
	maxImgSize = 5 << 20
)

func (h *InventoryHandler) HandleNew() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/inventory", http.StatusSeeOther)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxImgSize+1024)
		if err := r.ParseMultipartForm(maxImgSize); err != nil {
			fmt.Println("[Inventory]: Image File Too Large")
			// http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
			sse := datastar.NewSSE(w, r)
			sse.PatchElementTempl(general.MsgBox("Image Too Large", 3), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
			return
		}
		file, _, err := r.FormFile("image")
		if err != nil {
			// http.Error(w, "missing image", http.StatusBadRequest)
			sse := datastar.NewSSE(w, r)
			sse.PatchElementTempl(general.MsgBox("No Image Provided", 3), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
			return
		}
		defer file.Close()

		item := models.Item{
			OwnerID:     auth.GetUser(r.Context()).ID,
			Title:       r.FormValue("title"),
			Description: r.FormValue("description"),
		}

		item, err = addItem(h.database, file, item, h.uploadDir)
		if err != nil {
			fmt.Println(err)
			// http.Error(w, "invalid form data", http.StatusBadRequest)
			sse := datastar.NewSSE(w, r)
			sse.PatchElementTempl(general.MsgBox("Invalid Image Format", 3), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
			return
		}
		sse := datastar.NewSSE(w, r)
		// sse.PatchSignals([]byte(`{fileName: '', title: '', description: '', itemCount: 1}`))
		sse.PatchElementTempl(Item(item), datastar.WithSelectorID("item-list"), datastar.WithModeAppend())
		sse.PatchElementTempl(NewItemForm(), datastar.WithSelectorID("newItemForm"), datastar.WithModeReplace())
		sse.PatchElementTempl(general.MsgBox("Item Added", 1), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())
		sse.RemoveElementByID("new-item")
		sse.PatchElementTempl(NewItem(), datastar.WithSelectorID("item-list"), datastar.WithModeAppend())
	})
}

func (h *InventoryHandler) HandleDelete() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signals := &InventorySignals{}
		datastar.ReadSignals(r, signals)
		// fmt.Println("delete " + signals.SelectedItemID)
		if !isOwner(h.database, signals.SelectedItemID, auth.GetUser(r.Context())) {
			http.Error(w, "auth error", http.StatusUnauthorized)
			return
		}

		if err := deleteItem(h.database, signals.SelectedItemID, h.uploadDir); err != nil {
			fmt.Println("breaki")
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		sse := datastar.NewSSE(w, r)
		sse.PatchSignals([]byte(`{ showControls: false, selectedItem: ''}`))
		sse.RemoveElementByID("item-" + signals.SelectedItemID)
		sse.PatchElementTempl(general.MsgBox("Item Removed", 2), datastar.WithSelectorID("msg-box"), datastar.WithModePrepend())

	})
}

func (h *InventoryHandler) HandleSelect() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signals := &InventorySignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			fmt.Println(err)
			http.Error(w, "signals error", http.StatusInternalServerError)
			return
		}

		if !isOwner(h.database, signals.SelectedItemID, auth.GetUser(r.Context())) {
			http.Error(w, "auth error", http.StatusUnauthorized)
			return
		}

		item, err := getItem(h.database, signals.SelectedItemID)
		if err != nil {
			http.Error(w, "auth error", http.StatusUnauthorized)
			return
		}

		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(ItemContols(item))
		signals.ShowControls = true
		sse.MarshalAndPatchSignals(signals)
		// sse.PatchSignals([]byte(`{ showControls: true }`))
	})
}

func (h *InventoryHandler) HandleList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
		}
		signals := &InventorySignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			fmt.Println(err)
			http.Error(w, "signals error", http.StatusInternalServerError)
			return
		}

		if !isOwner(h.database, signals.SelectedItemID, auth.GetUser(r.Context())) {
			http.Error(w, "auth error", http.StatusUnauthorized)
			return
		}
		item, err := getItem(h.database, signals.SelectedItemID)
		if err != nil {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		sse := datastar.NewSSE(w, r)
		if err := listItem(h.database, item); err != nil {
			sse.PatchElementTempl(general.MsgBox("Error", 3), datastar.WithSelectorID("msg-box"), datastar.WithModeInner())
			return
		}
		item.Listed = true
		sse.PatchElementTempl(ItemContols(item))
		sse.PatchElementTempl(Item(item), datastar.WithSelectorID("item-"+item.ID))
		sse.PatchElementTempl(general.MsgBox("Item Listed", 1), datastar.WithSelectorID("msg-box"), datastar.WithModeInner())

	})
}
func (h *InventoryHandler) HandleDelist() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
		}
		signals := &InventorySignals{}
		if err := datastar.ReadSignals(r, signals); err != nil {
			fmt.Println(err)
			http.Error(w, "signals error", http.StatusInternalServerError)
			return
		}

		if !isOwner(h.database, signals.SelectedItemID, auth.GetUser(r.Context())) {
			http.Error(w, "auth error", http.StatusUnauthorized)
			return
		}
		item, err := getItem(h.database, signals.SelectedItemID)
		if err != nil {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		sse := datastar.NewSSE(w, r)
		if err := delistItem(h.database, item); err != nil {
			sse.PatchElementTempl(general.MsgBox("Error", 3), datastar.WithSelectorID("msg-box"), datastar.WithModeInner())
			return
		}
		item.Listed = false
		sse.PatchElementTempl(ItemContols(item))
		sse.PatchElementTempl(Item(item), datastar.WithSelectorID("item-"+item.ID))
		sse.PatchElementTempl(general.MsgBox("Item Delisted", 1), datastar.WithSelectorID("msg-box"), datastar.WithModeInner())

	})
}
