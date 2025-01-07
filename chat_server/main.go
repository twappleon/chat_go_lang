package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
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

		// 廣播訊息給所有其他客戶端
		clientsMut.Lock()
		for client := range clients {
			//if client != conn { // 不發送給發送者
			err := client.WriteMessage(websocket.TextMessage, rawMsg)
			if err != nil {
				log.Printf("Write error: %v", err)
				client.Close()
				delete(clients, client)
			}
			//}
		}
		clientsMut.Unlock()
	}
}

func handleGetChatHistory(w http.ResponseWriter, r *http.Request) {
	// Example response, modify as needed
	response := map[string]string{"message": "Chat history not implemented yet."}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/chat/history", handleGetChatHistory)
	log.Println("Server starting at :9999")
	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
