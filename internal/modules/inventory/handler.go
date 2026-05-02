package inventory

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/encador/trady/internal/models"
	"github.com/encador/trady/internal/modules/auth"
	"github.com/encador/trady/internal/templ/layout"
)

type InventoryHandler struct {
	database  *sql.DB
	imagesDir string
}

func NewHandler(db *sql.DB) *InventoryHandler {
	return &InventoryHandler{
		database:  db,
		imagesDir: "./images",
	}
}

func (h *InventoryHandler) InventoryPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opts := layout.Options{
			Content: InventoryPage(),
			URL:     "/inventory",
		}
		layout.Base(opts).Render(r.Context(), w)
	})
}

const (
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
			OwnerID: auth.GetUser(r.Context()).ID,
		}

		err = addItem(h.database, file, item, h.imagesDir)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "invalid form data", http.StatusBadRequest)
		}

	})
}
