package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type Server struct {
	db *pgxpool.Pool
}

func (s *Server) AddFriend(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id"`
		FriendID string `json:"friend_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}
	friendID, err := uuid.Parse(req.FriendID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend_id"})
		return
	}

	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	// Insert both directions to represent mutual friendship
	_, err = tx.Exec(ctx, `
        INSERT INTO friendships (user_id, friend_id) VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `, userID, friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert friendship"})
		return
	}
	_, err = tx.Exec(ctx, `
        INSERT INTO friendships (user_id, friend_id) VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `, friendID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert friendship"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friendship added"})
}

func (s *Server) GetFriends(c *gin.Context) {
	userIDStr := c.Param("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx := context.Background()
	rows, err := s.db.Query(ctx, `
        SELECT u.id, u.name, u.email
        FROM users u
        JOIN friendships f ON u.id = f.friend_id
        WHERE f.user_id = $1
    `, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch friends"})
		return
	}
	defer rows.Close()

	var friends []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user"})
			return
		}
		friends = append(friends, u)
	}

	c.JSON(http.StatusOK, friends)
}

func (s *Server) CheckFriend(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id"`
		FriendID string `json:"friend_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}
	friendID, err := uuid.Parse(req.FriendID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend_id"})
		return
	}

	ctx := context.Background()
	var exists bool
	err = s.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM friendships 
			WHERE user_id = $1 AND friend_id = $2
		)
	`, userID, friendID).Scan(&exists)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check friendship"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"are_friends": exists})
}

func main() {
	// Setup DB connection pool (adjust connection string)
	dbpool, err := pgxpool.New(context.Background(), "postgres://shaun:shaun@localhost:5432/wassupdb")
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer dbpool.Close()

	server := &Server{db: dbpool}
	r := gin.Default()

	r.POST("/addFriend", server.AddFriend)
	r.GET("/users/:userID/friends", server.GetFriends)
	r.POST("/checkFriends", server.CheckFriend)

	fmt.Println("Server running on :5002")
	r.Run(":5002")
}
