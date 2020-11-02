package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	http.HandleFunc("/signin", Signin)
	http.HandleFunc("/signup", Signup)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello! you've requested %s\n", r.URL.Path)
	})

	http.HandleFunc("/cached", func(w http.ResponseWriter, r *http.Request) {
		maxAgeParams, ok := r.URL.Query()["max-age"]
		if ok && len(maxAgeParams) > 0 {
			maxAge, _ := strconv.Atoi(maxAgeParams[0])
			w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", maxAge))
		}
		requestID := uuid.Must(uuid.NewV4())
		fmt.Fprintf(w, requestID.String())
	})

	http.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["key"]
		if ok && len(keys) > 0 {
			fmt.Fprintf(w, r.Header.Get(keys[0]))
			return
		}
		headers := []string{}
		for key, values := range r.Header {
			headers = append(headers, fmt.Sprintf("%s=%s", key, strings.Join(values, ",")))
		}
		fmt.Fprintf(w, strings.Join(headers, "\n"))
	})

	http.HandleFunc("/env", func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["key"]
		if ok && len(keys) > 0 {
			fmt.Fprintf(w, os.Getenv(keys[0]))
			return
		}
		envs := []string{}
		for _, env := range os.Environ() {
			envs = append(envs, env)
		}
		fmt.Fprintf(w, strings.Join(envs, "\n"))
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		codeParams, ok := r.URL.Query()["code"]
		if ok && len(codeParams) > 0 {
			statusCode, _ := strconv.Atoi(codeParams[0])
			if statusCode >= 200 && statusCode < 600 {
				w.WriteHeader(statusCode)
			}
		}
		requestID := uuid.Must(uuid.NewV4())
		fmt.Fprintf(w, requestID.String())
	})

	http.HandleFunc("/zip", func(w http.ResponseWriter, r *http.Request) {

		reqCity, ok := r.URL.Query()["city"]

		// variables
		var cities []city

		// open up database
		db, err := sql.Open("sqlite3", "./zipcode.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		searchCity := strings.Trim(reqCity[0], "\"")

		rows, err := db.Query("select zip, primaryCity, state, county, timezone, latitude, longitude, irsEstimatedPopulation2015 from zip_code_database where primaryCity like  '%' || ? || '%'  and type = 'STANDARD'", searchCity)

		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			city := city{}
			err = rows.Scan(&city.Zip, &city.City, &city.State, &city.County, &city.Timezone, &city.Latitude, &city.Longitude, &city.Population)
			cities = append(cities, city)
			if err != nil {
				log.Fatal(err)
			}
		}

		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

		if ok && len(cities) > 0 {

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cities)
		} else {
			w.WriteHeader(404)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	for _, encodedRoute := range strings.Split(os.Getenv("ROUTES"), ",") {
		if encodedRoute == "" {
			continue
		}
		pathAndBody := strings.SplitN(encodedRoute, "=", 2)
		path, body := pathAndBody[0], pathAndBody[1]
		http.HandleFunc("/"+path, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		})
	}

	bindAddr := fmt.Sprintf(":%s", port)
	lines := strings.Split(startupMessage, "\n")
	fmt.Println()
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println()
	fmt.Printf("==> Server listening at %s 🚀\n", bindAddr)

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
}

func initDB() {
	var err error

	db, err = sql.Open("sqlite3", "./bible-cal.db")
	if err != nil {
		panic(err)
	}
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("views/login")
		t.Execute(w, nil)
	} else {
		var err error

		creds := &Credentials{}

		err = json.NewDecoder(r.Body).Decode(creds)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)

		if _, err = db.Query("INSERT INTO USERS (username, email, password) VALUES ($1, $2, $3)", creds.UserName, creds.Email, string(hashedPassword)); err != nil {
			// If there is any issue with inserting into the database, return a 500 error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func Signin(w http.ResponseWriter, r *http.Request) {

}
