package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/encador/trady/internal/database"
	"github.com/encador/trady/internal/modules/users"
	"github.com/encador/trady/internal/templ/component"
)

type config struct {
	dbPath string
	init   bool
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// t := time.Now().Format("1-02 15:04:05")
		t := time.Now().Format("15:04:05")
		fmt.Printf("[%s] %s: %s\n", t, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func main() {

	var cnf config
	flag.StringVar(&cnf.dbPath, "db-path", "trady.db", "sqlite3 database file")
	flag.BoolVar(&cnf.init, "init", true, "initialize application files when missing")
	flag.Parse()

	if cnf.init {
		database.Create(cnf.dbPath)
		fmt.Println("[LOG] DB Created")
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
		// view.Base().Render(r.Context(), w)
		component.Hello("green").Render(r.Context(), w)
	})

	mux.Handle("/user", userH.HandleUserPage())
	mux.Handle("/user/new", userH.HandleAdd())

	fmt.Println("[LOG] Serving on localhost:555000")
	err = http.ListenAndServe("localhost:55000", logger(mux))
	fmt.Println(err)

}
