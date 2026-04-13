package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/encador/trady/internal/database"
	"github.com/encador/trady/internal/modules/users"
	"github.com/encador/trady/internal/templ/component"
)

type config struct {
	dbPath string
	init   bool
}

func main() {

	var cnf config
	flag.StringVar(&cnf.dbPath, "db-path", "trady.db", "sqlite3 database file")
	flag.BoolVar(&cnf.init, "init", true, "initialize application files when missing")
	flag.Parse()

	if cnf.init {
		database.Create(cnf.dbPath)
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
		fmt.Println(r.URL)
		// view.Base().Render(r.Context(), w)
		component.Hello("green").Render(r.Context(), w)
	})

	mux.Handle("/user", userH.HandleUserPage())

	fmt.Println("[LOG] Serving on localhost:555000")
	err = http.ListenAndServe("localhost:55000", mux)
	fmt.Println(err)

}
