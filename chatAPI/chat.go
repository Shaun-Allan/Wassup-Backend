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

type Message struct {
	Sender    string    `json:"sender" bson:"sender"`
	Receiver  string    `json:"receiver" bson:"receiver"`
	Content   string    `json:"content" bson:"content"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[string]*websocket.Conn)
var messagesCollection *mongo.Collection

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	messagesCollection = client.Database("wassupdb").Collection("dms")

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/history", getMessageHistory)

	fmt.Println("Server running on :5003")
	log.Fatal(http.ListenAndServe(":5003", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer ws.Close()
	clients[userID] = ws
	log.Println(userID, "connected")

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		msg.Timestamp = time.Now()
		go saveMessageToMongo(msg)

		if conn, ok := clients[msg.Receiver]; ok {
			conn.WriteJSON(msg)
		}
	}
	delete(clients, userID)
}

func saveMessageToMongo(msg Message) {
	filter := bson.M{"participants": bson.M{"$all": []string{msg.Sender, msg.Receiver}}}
	update := bson.M{"$push": bson.M{"messages": msg}, "$setOnInsert": bson.M{"participants": []string{msg.Sender, msg.Receiver}}}
	opts := options.Update().SetUpsert(true)
	messagesCollection.UpdateOne(context.TODO(), filter, update, opts)
}

func getMessageHistory(w http.ResponseWriter, r *http.Request) {
	user1 := r.URL.Query().Get("user1")
	user2 := r.URL.Query().Get("user2")

	filter := bson.M{"participants": bson.M{"$all": []string{user1, user2}}}
	var chat struct {
		Messages []Message `bson:"messages"`
	}

	err := messagesCollection.FindOne(context.TODO(), filter).Decode(&chat)
	if err != nil {
		json.NewEncoder(w).Encode([]Message{})
		return
	}
	json.NewEncoder(w).Encode(chat.Messages)
}
