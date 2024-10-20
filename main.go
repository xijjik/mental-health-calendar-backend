package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	_ "modernc.org/sqlite"
)

type Event struct {
	ID      int    `json:"id"`
	Date    string `json:"date"`
	Content string `json:"content"`
	Mood    string `json:"mood"`
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

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(r)

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
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
	rows, err := db.Query("SELECT id, date, content, mood FROM events ORDER BY date DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(&e.ID, &e.Date, &e.Content, &e.Mood)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		events = append(events, e)
	}

	json.NewEncoder(w).Encode(events)
}

func addEvent(w http.ResponseWriter, r *http.Request) {
	var e Event
	err := json.NewDecoder(r.Body).Decode(&e)
	if err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received event: %+v", e)

	dateStr := e.Date

	var existingID int
	err = db.QueryRow("SELECT id FROM events WHERE date = ?", dateStr).Scan(&existingID)
	
	if err == sql.ErrNoRows {
		result, err := db.Exec("INSERT INTO events (date, content, mood) VALUES (?, ?, ?)",
			dateStr, e.Content, e.Mood)
		if err != nil {
			http.Error(w, "Error inserting event: "+err.Error(), http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()
		e.ID = int(id)
	} else if err != nil {
		http.Error(w, "Error checking for existing event: "+err.Error(), http.StatusInternalServerError)
		return
	} else {
		_, err = db.Exec("UPDATE events SET content = ?, mood = ? WHERE id = ?",
			e.Content, e.Mood, existingID)
		if err != nil {
			http.Error(w, "Error updating event: "+err.Error(), http.StatusInternalServerError)
			return
		}
		e.ID = existingID
	}

	err = db.QueryRow("SELECT id, date, content, mood FROM events WHERE id = ?", e.ID).Scan(&e.ID, &dateStr, &e.Content, &e.Mood)
	if err != nil {
		http.Error(w, "Error retrieving event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Event to be returned: %+v", e)

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
		e.Date, e.Content, e.Mood, idInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	e.ID = idInt
	json.NewEncoder(w).Encode(e)
}
