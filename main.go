package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/encador/trady/internal/database"
	"github.com/encador/trady/internal/modules/middleware"
	"github.com/encador/trady/internal/modules/users"
	"github.com/encador/trady/internal/templ/component"
	"github.com/encador/trady/internal/templ/layout"
)

type config struct {
	address string
	port    int
	dbPath  string
	init    bool
}

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

	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		layout.Base(component.Hello("")).Render(r.Context(), w)
	})

	mux.Handle("/user", userH.HandleUserPage())
	mux.Handle("/user/new", userH.HandleAdd())
	mux.Handle("/user/login", userH.HandleLogin())
	mux.Handle("/user/logout", userH.HandleLogout())

	adr := fmt.Sprintf("%s:%d", cnf.address, cnf.port)

	fmt.Println("[LOG] Serving on " + adr)
	err = http.ListenAndServe(adr, middleware.AuthHandler(mux, db))
	fmt.Println(err)

}
