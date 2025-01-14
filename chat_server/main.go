package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"p2p_chat/config"
	"p2p_chat/database"
	"p2p_chat/models"
	"strconv"
	"strings"
	"sync"
	"time"

	"p2p_chat/lottery"

	"github.com/gin-gonic/gin"
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

func handleWebSocket(c *gin.Context) {
	w := c.Writer
	r := c.Request
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
		// Use the lottery package to generate winning numbers
		winningNumbers := lottery.GenerateWinningNumbers(rangeMax, selectCount)
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

func handleGetChatHistory(c *gin.Context) {
	userId := c.Query("userId")
	println("userId%s", userId)
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing userId"})
		return
	}

	messages, err := getMessages(userId, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	println(messages)

	// Initialize the response
	response := Response{
		Code:    int(StatusOK),
		Message: "Success",
		Data:    map[string]interface{}{"messages": messages},
	}
	c.JSON(http.StatusOK, response)
}

func handleCheckWinning(c *gin.Context) {
	userNumbersStr := c.Query("userNumbers")
	if userNumbersStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing userNumbers"})
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

	// Use the lottery package to generate winning numbers and check matches
	winningNumbers := lottery.GenerateWinningNumbers(49, 6)
	matchCount, matchedNumbers := lottery.CheckWinning(userNumbers, winningNumbers)

	data := map[string]interface{}{
		"winningNumbers": winningNumbers,
		"matchCount":     matchCount,
		"matchedNumbers": matchedNumbers,
	}
	response := Response{
		Code:    int(StatusOK),
		Message: string(MsgSuccess),
		Data:    map[string]interface{}{"data": data},
	}
	c.JSON(http.StatusOK, response)
}

func handleGenerateWinningNumbers(c *gin.Context) {
	// Use the lottery package to generate winning numbers
	winningNumbers := lottery.GenerateWinningNumbers(49, 6)

	// Prepare the response
	response := Response{
		Code:    int(StatusOK),
		Message: "Winning numbers generated successfully",
		Data:    map[string]interface{}{"winningNumbers": winningNumbers},
	}

	// Send the response as JSON
	c.JSON(http.StatusOK, response)
}

func handleCheckNumbers(c *gin.Context) {
	userNumbersStr := c.Query("userNumbers")
	if userNumbersStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing userNumbers"})
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

	// Use the lottery package to generate winning numbers and check matches
	winningNumbers := lottery.GenerateWinningNumbers(49, 6)
	matchCount, matchedNumbers := lottery.CheckWinning(userNumbers, winningNumbers)

	// Prepare the response
	response := Response{
		Code:    int(StatusOK),
		Message: "Check completed successfully",
		Data:    map[string]interface{}{"winningNumbers": winningNumbers, "matchCount": matchCount, "matchedNumbers": matchedNumbers},
	}

	// Send the response as JSON
	c.JSON(http.StatusOK, response)
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

	r := gin.Default()

	// 使用 Gin 路由器定义 API 路由
	api := r.Group("/api")
	{
		api.GET("/ws", handleWebSocket)
		api.GET("/chat/history", handleGetChatHistory)
		api.GET("/check-winning", handleCheckWinning)
		api.GET("/generate-winning-numbers", handleGenerateWinningNumbers)
		api.GET("/check-number", handleCheckNumbers)
	}

	log.Println("Server starting at :8888")
	if err := r.Run(":8888"); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
