package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB

type UserResponse struct {
	Name string `json:"name"`
}

func main() {
	var err error

	connStr := "postgres://shaun:shaun@localhost/wassupdb?sslmode=disable"
	db, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/get-name", getNameHandler)
	fmt.Println("Server started on :5006")
	log.Fatal(http.ListenAndServe(":5006", nil))
}

func getNameHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	var name string
	err := db.QueryRowContext(context.Background(), "SELECT name FROM users WHERE id = $1", userID).Scan(&name)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserResponse{Name: name})
}
