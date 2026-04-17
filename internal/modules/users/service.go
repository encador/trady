package users

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/encador/trady/internal/models"
)

func addUser(user models.User, db *sql.DB) []string {
	errs := []string{}

	// Validate User Input
	if len(strings.Fields(user.Username)) > 1 {
		errs = append(errs, "Name: Must be a Single Word")
	}

	username := strings.TrimSpace(user.Username)
	if username == "" {
		errs = append(errs, "Name: Connot be Empty")
	} else if len(username) > 24 {
		errs = append(errs, "Name: Connot Exceed 24 Characters")
	}

	if user.Password == "" {
		errs = append(errs, "Password: Connot be Empty")
	} else if len(user.Password) > 64 {
		errs = append(errs, "Password: Connot Exceed 64 Characters")
	}
	if len(errs) != 0 {
		return errs
	}

	// Add User to Database
	q := `insert into users(username, password) values(?,?)`
	if _, err := db.Exec(q, username, user.Password); err == nil {
		errs = append(errs, "Success")
	} else {
		fmt.Println(err)
		errs = append(errs, "Name: Already Taken")
		return errs
	}

	return errs
}
