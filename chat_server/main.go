package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"p2p_chat/config"
	"p2p_chat/database"
	"p2p_chat/models"
	"sort"
	"strconv"
	"strings"
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

type LotteryResponse struct {
	winningNumbers []int `json:"winning_nunbers"`
}

type MatchResponse struct {
	MatchedNumber []int `json:"matched_numbers"`
	MatchCount    int   `json:"match_count"`
}

type Code int

const (
	StatusOK         Code = 200
	StatusNotFound   Code = 404
	BalanceNotEnough Code = 2003
)

type StatusMessage string

const (
	MsgSuccess          StatusMessage = "成功"
	MsgNotFound         StatusMessage = "找不到"
	MsgBalanceNotEnough StatusMessage = "餘額不足"
)

type Response struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func saveMessage(msg *models.ChatMessage) error {
	collection := database.ChatDB.Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg.ID = primitive.NewObjectID()
	msg.Timestamp = time.Now()
	log.Printf("---insert---: %v", msg.Sender)
	result, err := collection.InsertOne(ctx, msg)
	if err != nil {
		fmt.Println("Error inserting document:", err)
		return err
	}

	// 检查返回的插入结果
	log.Printf("Insert successful! Inserted ID: %v", result.InsertedID)
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

		// 遊戲設定
		rangeMax := 49
		selectCount := 6
		// 生成開獎號碼
		winningNumbers := generateWinningNumbers(rangeMax, selectCount)
		fmt.Printf("開獎號碼: %v\n", winningNumbers)

		// Save the message to the database
		chatMessage := &models.ChatMessage{
			Sender:  msg.Sender,
			Message: fmt.Sprintf("%v", winningNumbers),
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

	// Initialize the response
	response := Response{
		Code:    int(StatusOK),
		Message: "Success",
		Data:    map[string]interface{}{"messages": messages},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 生成開獎號碼
func generateWinningNumbers(rangeMax, count int) []int {
	rand.Seed(time.Now().UnixNano())
	numbers := rand.Perm(rangeMax)[:count]
	sort.Ints(numbers) // 排序，方便匹配
	return numbers
}

// 全部組合 (n選k)
func combinations(numbers []int, k int) [][]int {
	var result [][]int
	var comb func(start int, combo []int)

	comb = func(start int, combo []int) {
		if len(combo) == k {
			// 建立一個副本以避免共用切片
			comboCopy := append([]int{}, combo...)
			result = append(result, comboCopy)
			return
		}

		for i := start; i < len(numbers); i++ {
			comb(i+1, append(combo, numbers[i]))
		}
	}

	comb(0, []int{})
	return result
}

// 比對結果
func checkWinning(userNumbers, winningNumbers []int) int {
	matchCount := 0
	for _, num := range userNumbers {
		for _, winNum := range winningNumbers {
			if num == winNum {
				matchCount++
			}
		}
	}
	return matchCount
}

func handleCheckWinning(w http.ResponseWriter, r *http.Request) {
	userNumbersStr := r.URL.Query().Get("userNumbers")
	if userNumbersStr == "" {
		http.Error(w, "Missing userNumbers", http.StatusBadRequest)
		return
	}

	// Convert userNumbersStr to a slice of integers
	var userNumbers []int
	// Assuming userNumbers are passed as a comma-separated string
	for _, numStr := range strings.Split(userNumbersStr, ",") {
		num, err := strconv.Atoi(numStr)
		if err == nil {
			userNumbers = append(userNumbers, num)
		}
	}

	winningNumbers := generateWinningNumbers(49, 6) // Example: generate 6 winning numbers from 1 to 49
	matchCount := checkWinning(userNumbers, winningNumbers)

	data := map[string]interface{}{
		"winningNumbers": winningNumbers,
		"matchCount":     matchCount,
	}
	response := Response{
		Code:    int(StatusOK),
		Message: string(MsgSuccess),
		Data:    map[string]interface{}{"data": data},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGenerateWinningNumbers(w http.ResponseWriter, r *http.Request) {
	// Generate winning numbers
	winningNumbers := generateWinningNumbers(49, 6) // Example: generate 6 winning numbers from 1 to 49

	// Prepare the response
	response := Response{
		Code:    int(StatusOK),
		Message: "Winning numbers generated successfully",
		Data:    map[string]interface{}{"winningNumbers": winningNumbers},
	}

	// Set the response header and encode the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCheckNumbers(w http.ResponseWriter, r *http.Request) {
	userNumbersStr := r.URL.Query().Get("userNumbers")
	if userNumbersStr == "" {
		http.Error(w, "Missing userNumbers", http.StatusBadRequest)
		return
	}

	// Convert userNumbersStr to a slice of integers
	var userNumbers []int
	for _, numStr := range strings.Split(userNumbersStr, ",") {
		num, err := strconv.Atoi(numStr)
		if err == nil {
			userNumbers = append(userNumbers, num)
		}
	}

	// Generate winning numbers for comparison
	winningNumbers := generateWinningNumbers(49, 6) // Example: generate 6 winning numbers from 1 to 49
	matchCount := checkWinning(userNumbers, winningNumbers)

	// Prepare the response
	response := Response{
		Code:    int(StatusOK),
		Message: "Check completed successfully",
		Data:    map[string]interface{}{"winningNumbers": winningNumbers, "matchCount": matchCount},
	}

	// Set the response header and encode the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// 載入配置
	cfg := config.LoadConfig()

	// 連接MongoDB
	if err := database.ConnectMongoDB(cfg); err != nil {
		log.Fatal(err)
	}
	defer database.CloseMongoDB()

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/chat/history", handleGetChatHistory)
	http.HandleFunc("/api/check-winning", handleCheckWinning)
	http.HandleFunc("/api/handleGenerateWinningNumbers", handleGenerateWinningNumbers)
	http.HandleFunc("/api/checkNumber", handleCheckNumbers)
	log.Println("Server starting at :8888")
	if err := http.ListenAndServe(":8888", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
