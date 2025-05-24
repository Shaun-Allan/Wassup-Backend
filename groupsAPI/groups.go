package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	db                    *sql.DB
	mongoClient           *mongo.Client
	mongoGroupsCollection *mongo.Collection
)

type Group struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
}

type AddMemberRequest struct {
	Members []uuid.UUID `json:"members"`
}

func main() {
	// Connect to PostgreSQL
	var err error
	db, err = sql.Open("pgx", "postgres://shaun:shaun@localhost:5432/wassupdb")
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer db.Close()

	// Connect to MongoDB
	mongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongoClient.Disconnect(context.TODO())

	mongoGroupsCollection = mongoClient.Database("wassupdb").Collection("groups_meta")

	// Setup routes
	r := mux.NewRouter()
	r.HandleFunc("/createGroup", createGroupHandler).Methods("POST")
	r.HandleFunc("/groups/{id}/addMember", addMemberHandler).Methods("POST")
	r.HandleFunc("/users/{id}/groups", getUserGroupsHandler).Methods("GET")

	fmt.Println("Server started at :5004")
	log.Fatal(http.ListenAndServe(":5004", r))
}

func createGroupHandler(w http.ResponseWriter, r *http.Request) {
	var g Group
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if g.Name == "" {
		http.Error(w, "Group name is required", http.StatusBadRequest)
		return
	}

	// Generate UUID
	g.ID = uuid.New()

	// Insert into PostgreSQL
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO groups (id, name, description) VALUES ($1, $2, $3)",
		g.ID, g.Name, g.Description)
	if err != nil {
		log.Println("PostgreSQL insert error:", err)
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	// Insert into MongoDB
	mongoDoc := bson.M{
		"_id":      g.ID.String(),
		"name":     g.Name,
		"members":  []string{},
		"messages": []bson.M{},
	}
	_, err = mongoGroupsCollection.InsertOne(context.TODO(), mongoDoc)
	if err != nil {
		log.Println("MongoDB insert warning:", err) // Continue even if this fails
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(g)
}

func addMemberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Members) == 0 {
		http.Error(w, "No members provided", http.StatusBadRequest)
		return
	}

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(context.Background(), `
		INSERT INTO group_memberships (group_id, user_id)
		VALUES ($1, $2) ON CONFLICT DO NOTHING`)
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	var memberStrs []string
	for _, userID := range req.Members {
		_, err := stmt.ExecContext(context.Background(), groupID, userID)
		if err != nil {
			http.Error(w, "Failed to add member", http.StatusInternalServerError)
			return
		}
		memberStrs = append(memberStrs, userID.String())
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Update MongoDB
	_, err = mongoGroupsCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": groupID.String()},
		bson.M{"$addToSet": bson.M{"members": bson.M{"$each": memberStrs}}},
	)
	if err != nil {
		log.Println("MongoDB update warning:", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func getUserGroupsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT g.id, g.name, g.description
		FROM groups g
		JOIN group_memberships gm ON g.id = gm.group_id
		WHERE gm.user_id = $1
	`

	rows, err := db.QueryContext(context.Background(), query, userID)
	if err != nil {
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Description); err != nil {
			http.Error(w, "Failed to parse group", http.StatusInternalServerError)
			return
		}
		groups = append(groups, g)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}
