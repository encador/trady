package users

import (
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/encador/trady/internal/models"
)

// military grade encryption :)
func HashPass(pass string, user string) string {
	data := []byte("superDUPERs3cure" + pass + "ufshi8H8()#)sudfh3484*$*#8" + user)
	h := sha512.Sum512(data)
	return hex.EncodeToString(h[:])
}

func SetRedirectCookie(url string, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "redirect",
		Value:    url,
		Path:     "/user",
		Expires:  time.Now().Add(time.Minute * 5),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func UnsetRedirectCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "redirect",
		Value:    "",
		Path:     "/user",
		Expires:  time.Now(),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func getRedirectURL(r *http.Request) (string, error) {
	cookie, err := r.Cookie("redirect")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func GetUser(username string, db *sql.DB) (models.User, error) {
	q := `select id, security, username, password from users where username = ?`
	row := db.QueryRow(q, username)
	user := models.User{}
	err := row.Scan(&user.ID, &user.Security, &user.Username, &user.Password)
	return user, err
}

func verifyPass(username, password string, db *sql.DB) error {
	u, err := GetUser(username, db)
	if err != nil {
		fmt.Println(err)
		return errors.New("[verifyPass] Wrong Username")
	}
	if username != u.Username || HashPass(password, username) != u.Password {
		return errors.New("[verifyPass] Wrong Password")
	}

	return nil
}

func addUser(user models.User, db *sql.DB) ([]string, error) {
	msgs := []string{}

	// Validate User Input
	if len(strings.Fields(user.Username)) > 1 {
		msgs = append(msgs, "Name: Must be a Single Word")
	}

	username := strings.TrimSpace(user.Username)
	if username == "" {
		msgs = append(msgs, "Name: Cannot be Empty")
	} else if len(username) > 24 {
		msgs = append(msgs, "Name: Cannot Exceed 24 Characters")
	}

	if user.Password == "" {
		msgs = append(msgs, "Password: Cannot be Empty")
	} else if len(user.Password) > 64 {
		msgs = append(msgs, "Password: Cannot Exceed 64 Characters")
	}
	if len(msgs) != 0 {
		return msgs, errors.New("[addUser]: Invalid Input")
	}

	// Add User to Database
	q := `insert into users(username, password) values(?,?)`
	if _, err := db.Exec(q, username, HashPass(user.Password, username)); err == nil {
		m := fmt.Sprintf("User %s Created", username)
		msgs = append(msgs, m)
	} else {
		fmt.Println(err)
		msgs = append(msgs, "Name: Already Taken")
		return msgs, fmt.Errorf("[addUser]: Name %s Already Taken", username)
	}

	return msgs, nil
}
