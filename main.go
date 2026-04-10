package main

import (
	"fmt"

	"github.com/encador/trady/internal"
	"github.com/encador/trady/internal/database"
)

func main() {
	internal.Hello()
	err := database.Create("bin/trady.db")
	fmt.Println(err)
	db, err := database.Open("bin/trady.db")
	if err == nil {
		defer func (){
			db.Close()
			fmt.Println("[LOG] DB Closed")
		}()
	} else {
		fmt.Println(err)
	}
}
