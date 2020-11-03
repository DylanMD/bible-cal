package data

import (
	"fmt"
	"net/http"

	"cal.bible/controllers"
	"golang.org/x/crypto/bcrypt"
)

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		controllers.Render(w, "../views/login.html", nil)
	} else {
		r.ParseForm()

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), 8)

		_, err := Db.Query("INSERT INTO USERS (username,email,password) VALUES (?,?,?)", r.FormValue("username"), r.FormValue("email"), string(hashedPassword))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf(err.Error())
		}
	}
}

func Signin(w http.ResponseWriter, r *http.Request) {

}
