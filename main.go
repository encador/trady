package main

import (
	"embed"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/encador/trady/internal/database"
	"github.com/encador/trady/internal/modules/middleware"
	"github.com/encador/trady/internal/modules/users"
	"github.com/encador/trady/internal/modules/inventory"
	"github.com/encador/trady/internal/templ/component"
	"github.com/encador/trady/internal/templ/layout"
)

type config struct {
	address string
	port    int
	dbPath  string
	init    bool
}

//go:embed static/*
var staticFiles embed.FS

func main() {

	var cnf config
	flag.StringVar(&cnf.address, "address", "localhost", "address on which the application runs")
	flag.IntVar(&cnf.port, "port", 55000, "Port # for the application")
	flag.StringVar(&cnf.dbPath, "db-path", "trady.db", "sqlite3 database file")
	flag.BoolVar(&cnf.init, "init", true, "initialize application files when missing")
	flag.Parse()

	if cnf.init {
		err := database.Create(cnf.dbPath)
		if err == nil {
			fmt.Println("[LOG] DB Created")
		}
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
	invH := inventory.NewHandler(db)

	fs := http.FileServer(http.FS(staticFiles))
	mux.Handle("/static/", middleware.Cache24(fs))
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static/robots.txt"
		fs.ServeHTTP(w, r)
	})


	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		layout.Base(layout.Options{Content: component.Hello(""), URL: "/"}).Render(r.Context(), w)
	})

	mux.Handle("/user", userH.HandleUserPage())
	mux.Handle("/user/new", userH.HandleAdd())
	mux.Handle("/user/login", userH.HandleLoginPage())
	mux.Handle("/user/logout", userH.HandleLogout())

	mux.Handle("/inventory", invH.InventoryPage())
	mux.Handle("/inventory/new", invH.HandleNew())
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	adr := fmt.Sprintf("%s:%d", cnf.address, cnf.port)

	fmt.Println("[LOG] Serving on " + adr)
	err = http.ListenAndServe(adr, middleware.AuthHandler(mux, db))
	fmt.Println(err)

}
