package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Data struct {
	ID       int
	Dist     float64 `json:"dist"`
	For100Km float64 `json:"for100Km"`
	Price    float64 `json:"price"`
}

func main() {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS petrol (id serial PRIMARY KEY ,dist REAL,for100Km REAL, price REAL)")
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/use", getData(db)).Methods("GET")
	router.HandleFunc("/use", fuelUse(db)).Methods("POST")
	router.HandleFunc("/use/{id}", delete(db)).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8000", jsonContent(router)))
}

func jsonContent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func getData(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM petrol")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		data := []Data{}

		for rows.Next() {
			var u Data
			if err := rows.Scan(&u.ID, &u.Dist, &u.For100Km, &u.Price); err != nil {
				log.Fatal(err)
			}
			data = append(data, u)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(data)
	}
}

func fuelUse(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var p Data
		json.NewDecoder(r.Body).Decode(&p)

		err := db.QueryRow("INSERT INTO petrol (dist, for100km, price) VALUES($1,$2,$3) RETURNING id", p.Dist, p.For100Km, p.Price).Scan(&p.ID)
		if err != nil {
			log.Fatal(err)
		}
		sum := ((p.Dist * p.For100Km) / 100) * p.Price
		json.NewEncoder(w).Encode(sum)

	}
}

func delete(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var p Data

		err := db.QueryRow("SELECT * FROM petrol WHERE id=$1", id).Scan(&p.ID, &p.Dist, &p.For100Km, &p.Price)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := db.Exec("DELETE FROM petrol WHERE id=$1", id)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			}
			json.NewEncoder(w).Encode("Deleted")
		}

	}
}
