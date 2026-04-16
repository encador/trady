package users

import (
	"database/sql"
	"fmt"

	"github.com/encador/trady/internal/models"
)

func addUser(user models.User, db *sql.DB) []string {
	errs := []string{}

	// Validate User Input
	if user.Username == "" || user.Username == "bob" {
		errs = append(errs, "Name: Connot Be Empty")
	}
	if user.Password == "" {
		errs = append(errs, "Password: Connot Be Empty")
	}
	if len(errs) != 0 {
		return errs
	}

	// Add User to Database
	q := `insert into users(username, password) values(?,?)`
	if _, err := db.Exec(q, user.Username, user.Password); err == nil {
		errs = append(errs, "Success")
	} else {
		fmt.Println(err)
		errs = append(errs, "Name: Already Taken")
		return errs
	}

	return errs
}
