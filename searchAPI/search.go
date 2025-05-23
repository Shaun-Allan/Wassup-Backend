package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SearchRequest struct {
	Query string `json:"query"`
}

var db *pgxpool.Pool

func main() {
	ctx := context.Background()

	dbUrl := "postgres://shaun:shaun@localhost:5432/wassupdb"
	var err error
	db, err = pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	router := gin.Default()
	router.POST("/search", searchUsers)

	router.Run(":5001")
}

func searchUsers(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	query := "%" + req.Query + "%"
	rows, err := db.Query(context.Background(),
		`SELECT id, name, email, password FROM users WHERE name ILIKE $1 OR email ILIKE $1`, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Password); err != nil {
			continue
		}
		users = append(users, u)
	}

	c.JSON(http.StatusOK, users)
}
