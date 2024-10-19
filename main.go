package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

type Event struct {
	ID      int       `json:"id"`
	Date    time.Time `json:"date"`
	Content string    `json:"content"`
	Mood    string    `json:"mood"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "./calendar.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()

	r := mux.NewRouter()
	r.HandleFunc("/events", getEvents).Methods("GET")
	r.HandleFunc("/events", addEvent).Methods("POST")
	r.HandleFunc("/events/{id}", updateEvent).Methods("PUT")

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date DATE,
		content TEXT,
		mood TEXT
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, date, content, mood FROM events")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var dateStr string
		err := rows.Scan(&e.ID, &dateStr, &e.Content, &e.Mood)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		e.Date, _ = time.Parse("2006-01-02", dateStr)
		events = append(events, e)
	}

	json.NewEncoder(w).Encode(events)
}

func addEvent(w http.ResponseWriter, r *http.Request) {
	var e Event
	err := json.NewDecoder(r.Body).Decode(&e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO events (date, content, mood) VALUES (?, ?, ?)",
		e.Date.Format("2006-01-02"), e.Content, e.Mood)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	e.ID = int(id)
	json.NewEncoder(w).Encode(e)
}

func updateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var e Event
	err := json.NewDecoder(r.Body).Decode(&e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE events SET date = ?, content = ?, mood = ? WHERE id = ?",
		e.Date.Format("2006-01-02"), e.Content, e.Mood, idInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	e.ID = idInt
	json.NewEncoder(w).Encode(e)
}
