package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"` // do not return password in API response
}

func main() {
	var err error
	db, err = pgxpool.New(context.Background(), "postgres://shaun:shaun@localhost:5432/wassupdb")
	if err != nil {
		log.Fatalf("Unable to connect to DB: %v\n", err)
	}
	defer db.Close()

	r := gin.Default()

	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)

	r.Run(":5000")
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func registerHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Store password as plain text (not recommended for production)
	query := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`
	var user User
	err := db.QueryRow(context.Background(), query, req.Name, req.Email, req.Password).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User creation failed (maybe email exists)"})
		return
	}

	user.Name = req.Name
	user.Email = req.Email
	user.Password = "" // never return password

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "user": user})
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	query := `SELECT id, name, email, password FROM users WHERE email=$1`
	err := db.QueryRow(context.Background(), query, req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password dkjf"})
		return
	}

	if user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	user.Password = "" // don't send password back

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": user})
}
