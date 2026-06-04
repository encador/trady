package main

import (
	"embed"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/encador/trady/internal/database"
	"github.com/encador/trady/internal/inventory"
	"github.com/encador/trady/internal/middleware"
	"github.com/encador/trady/internal/users"
	"github.com/encador/trady/internal/general"
)

type config struct {
	address   string
	port      int
	dbPath    string
	init      bool
	uploadDir string
}

//go:embed static/*
var staticFiles embed.FS

func main() {

	var cnf config
	flag.StringVar(&cnf.address, "address", "localhost", "address on which the application runs")
	flag.IntVar(&cnf.port, "port", 55000, "Port # for the application")
	flag.StringVar(&cnf.dbPath, "db", "trady.db", "sqlite3 database file")
	flag.BoolVar(&cnf.init, "init", true, "initialize application files when missing")
	flag.StringVar(&cnf.uploadDir, "uploads", "./uploads", "directory for user uploaded item-images")
	flag.Parse()

	if cnf.init {
		err := database.Create(cnf.dbPath)
		if err == nil {
			fmt.Println("[LOG] DB Created")
		}
		err = os.MkdirAll(cnf.uploadDir, 0755)
		if err != nil {
			fmt.Println(err)
			fmt.Println("[ERROR] UploadDir Not Created")
			os.Exit(0)
		}
		fmt.Println("[LOG] UploadDir Verified")
	}

	db, err := database.Open(cnf.dbPath)
	if err == nil {
		fmt.Println("[LOG] DB Opened")
		defer func() { // Does not run when using ctr-c to close
			db.Close()
			fmt.Println("[LOG] DB Closed")
		}()
	} else {
		fmt.Println(err)
		os.Exit(0)
	}

	mux := http.NewServeMux()
	userH := users.NewHandler(db)
	invH, err := inventory.NewHandler(db, cnf.uploadDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	fs := http.FileServer(http.FS(staticFiles))
	mux.Handle("/static/", middleware.Cache1(fs))
	// mux.Handle("/static/",fs)
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static/robots.txt"
		fs.ServeHTTP(w, r)
	})

	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		general.Base(general.Options{Content: general.Hello(""), URL: "/"}).Render(r.Context(), w)
	})

	mux.Handle("/user", userH.HandleUserPage())
	mux.Handle("/user/new", userH.HandleAdd())
	mux.Handle("/user/login", userH.HandleLoginPage())
	mux.Handle("/user/logout", userH.HandleLogout())

	mux.Handle("/inventory", invH.InventoryPage())
	mux.Handle("/inventory/new", invH.HandleNew())
	mux.Handle("/inventory/select", invH.HandleSelect())
	mux.Handle("/inventory/delete", invH.HandleDelete())
	mux.Handle("/inventory/list", invH.HandleList())

	mux.Handle("/images/", http.StripPrefix("/images/", middleware.Cache1(http.FileServer(http.Dir(cnf.uploadDir)))))
	// mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(cnf.uploadDir))))

	adr := fmt.Sprintf("%s:%d", cnf.address, cnf.port)

	fmt.Println("[LOG] Serving on " + adr)
	err = http.ListenAndServe(adr, middleware.AuthHandler(mux, db))
	fmt.Println(err)

}
