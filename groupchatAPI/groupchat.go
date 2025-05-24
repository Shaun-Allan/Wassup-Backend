package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GroupMessage struct {
	GroupID   string    `json:"group_id" bson:"group_id"`
	Sender    string    `json:"sender" bson:"sender"`
	Content   string    `json:"content" bson:"content"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var groupClients = make(map[string]map[string]*websocket.Conn) // groupID -> userID -> conn
var groupMessagesCollection *mongo.Collection

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	groupMessagesCollection = client.Database("wassupdb").Collection("groups_meta")

	http.HandleFunc("/group/ws", handleGroupWebSocket)
	http.HandleFunc("/group/history", getGroupMessageHistory)

	fmt.Println("Group server running on :5005")
	log.Fatal(http.ListenAndServe(":5005", nil))
}

func handleGroupWebSocket(w http.ResponseWriter, r *http.Request) {
	groupID := r.URL.Query().Get("group_id")
	userID := r.URL.Query().Get("user_id")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer ws.Close()

	// Create group if not exists
	if groupClients[groupID] == nil {
		groupClients[groupID] = make(map[string]*websocket.Conn)
	}
	groupClients[groupID][userID] = ws

	log.Println("User", userID, "joined group", groupID)

	for {
		var msg GroupMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		msg.Timestamp = time.Now()

		go saveGroupMessageToMongo(msg)

		// Broadcast to all users in group
		for uid, conn := range groupClients[groupID] {
			if uid != msg.Sender { // Don't send back to sender
				conn.WriteJSON(msg)
			}
		}
	}

	delete(groupClients[groupID], userID)
}

func saveGroupMessageToMongo(msg GroupMessage) {
	filter := bson.M{"_id": msg.GroupID}
	update := bson.M{
		"$push":        bson.M{"messages": msg},
		"$setOnInsert": bson.M{"members": []string{msg.Sender}}, // optional: keep track of members
	}
	opts := options.Update().SetUpsert(true)
	groupMessagesCollection.UpdateOne(context.TODO(), filter, update, opts)
}

func getGroupMessageHistory(w http.ResponseWriter, r *http.Request) {
	groupID := r.URL.Query().Get("group_id")

	filter := bson.M{"_id": groupID}
	var group struct {
		Messages []GroupMessage `bson:"messages"`
	}

	err := groupMessagesCollection.FindOne(context.TODO(), filter).Decode(&group)
	if err != nil {
		log.Println("Group not found or no messages")
		json.NewEncoder(w).Encode([]GroupMessage{})
		return
	}
	json.NewEncoder(w).Encode(group.Messages)
}
