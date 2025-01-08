package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"p2p_chat/database"
	"p2p_chat/models"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients    = make(map[*websocket.Conn]bool)
	clientsMut sync.Mutex
)

type Message struct {
	Type      string                 `json:"type"`
	Action    string                 `json:"action,omitempty"`
	Sender    string                 `json:"sender,omitempty"`
	Message   string                 `json:"message,omitempty"`
	SDP       map[string]interface{} `json:"sdp,omitempty"`
	Candidate map[string]interface{} `json:"candidate,omitempty"`
}

func saveMessage(msg *models.ChatMessage) error {
	collection := database.ChatDB.Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg.ID = primitive.NewObjectID()
	msg.Timestamp = time.Now()
	println("---insert---", msg.Sender)
	result, err := collection.InsertOne(ctx, msg)
	if err != nil {
		fmt.Println("Error inserting document:", err)
		return err
	}

	// 检查返回的插入结果
	fmt.Println("Insert successful! Inserted ID:", result.InsertedID)
	return err
}

func getMessages(userId string, limit int64) ([]models.ChatMessage, error) {
	collection := database.ChatDB.Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 查詢與用戶相關的消息
	filter := bson.M{
		"$or": []bson.M{
			{"sender": userId},
			{"receiver": userId},
		},
	}

	opts := options.Find().
		SetSort(bson.D{{"timestamp", -1}}).
		SetLimit(limit)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.ChatMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	println("messages:$s", len(messages))

	return messages, nil
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Websocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	clientsMut.Lock()
	clients[conn] = true
	clientsMut.Unlock()

	defer func() {
		clientsMut.Lock()
		delete(clients, conn)
		clientsMut.Unlock()
	}()

	for {
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		var msg Message
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			log.Printf("JSON parse error: %v", err)
			continue
		}

		// Save the message to the database
		chatMessage := &models.ChatMessage{
			Sender:  msg.Sender,
			Message: msg.Message,
			// Add other fields as necessary
		}

		println(chatMessage.Message)
		if err := saveMessage(chatMessage); err != nil {
			log.Printf("Error saving message: %v", err)
		}

		// 廣播訊息給所有其他客戶端
		clientsMut.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, rawMsg)
			if err != nil {
				log.Printf("Write error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		clientsMut.Unlock()
	}
}

func handleGetChatHistory(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	println("userId%s", userId)
	if userId == "" {
		http.Error(w, "Missing userId", http.StatusBadRequest)
		return
	}

	messages, err := getMessages(userId, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	println(messages)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func main() {
	// 連接MongoDB
	if err := database.ConnectMongoDB(); err != nil {
		log.Fatal(err)
	}
	defer database.CloseMongoDB()

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/chat/history", handleGetChatHistory)
	log.Println("Server starting at :8888")
	if err := http.ListenAndServe(":8888", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
