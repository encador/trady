package users

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/encador/trady/internal/models"
)

func addUser(user models.User, db *sql.DB) ([]string, error) {
	msgs := []string{}

	// Validate User Input
	if len(strings.Fields(user.Username)) > 1 {
		msgs = append(msgs, "Name: Must be a Single Word")
	}

	username := strings.TrimSpace(user.Username)
	if username == "" {
		msgs = append(msgs, "Name: Connot be Empty")
	} else if len(username) > 24 {
		msgs = append(msgs, "Name: Connot Exceed 24 Characters")
	}

	if user.Password == "" {
		msgs = append(msgs, "Password: Connot be Empty")
	} else if len(user.Password) > 64 {
		msgs = append(msgs, "Password: Connot Exceed 64 Characters")
	}
	if len(msgs) != 0 {
		return msgs, errors.New("[addUser]: Invalid Input")
	}

	// Add User to Database
	q := `insert into users(username, password) values(?,?)`
	if _, err := db.Exec(q, username, user.Password); err == nil {
		m := fmt.Sprintf("User %s Created", username)
		msgs = append(msgs, m)
	} else {
		fmt.Println(err)
		msgs = append(msgs, "Name: Already Taken")
		return msgs, fmt.Errorf("[addUser]: Name %s Already Taken", username)
	}

	return msgs, nil
}
