package inventory

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/encador/trady/internal/models"
	"github.com/encador/trady/internal/modules/auth"
	"github.com/encador/trady/internal/templ/layout"
	"github.com/starfederation/datastar-go/datastar"
)

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
		opts := layout.Options{
			Content:  InventoryPage(items),
			Contorls: InventoryControl(),
			URL:      "/inventory",
		}
		layout.Base(opts).Render(r.Context(), w)
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
			http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
			fmt.Println("[Inventory]: Image File Too Large")
			return
		}
		file, _, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "missing image", http.StatusBadRequest)
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
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(Item(item), datastar.WithSelectorID("item-list"), datastar.WithModeAppend())
		sse.PatchElementTempl(NewItemForm(), datastar.WithModeReplace(), datastar.WithSelectorID("form-container"))
		sse.PatchSignals([]byte(`{fileName: '', title: '', description: '', itemCount: 1}`))

	})
}
