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
	"sort"
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

var requestPool = sync.Pool{
	New: func() interface{} { return &Response{} },
}

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

type Friend struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var friends = make(map[string][]Friend) // 存储用户的好友列表

type Order struct {
	ID     int
	Type   string // "buy" or "sell"
	Price  float64
	Amount int
}

type OrderBook struct {
	BuyOrders  []Order
	SellOrders []Order
}

func (ob *OrderBook) AddOrder(order Order) error {
	// 驗證訂單類型
	if order.Type != "buy" && order.Type != "sell" {
		return fmt.Errorf("invalid order type: %s", order.Type)
	}

	if order.Type == "buy" {
		ob.BuyOrders = append(ob.BuyOrders, order)
		// 對買單按價格排序（從高到低）
		sort.Slice(ob.BuyOrders, func(i, j int) bool {
			return ob.BuyOrders[i].Price > ob.BuyOrders[j].Price
		})
	} else {
		ob.SellOrders = append(ob.SellOrders, order)
		// 對賣單按價格排序（從低到高）
		sort.Slice(ob.SellOrders, func(i, j int) bool {
			return ob.SellOrders[i].Price < ob.SellOrders[j].Price
		})
	}
	return nil
}

func (ob *OrderBook) MatchOrders() {
	fmt.Printf("--MatchOrders--%d\n", len(ob.BuyOrders))
	for len(ob.BuyOrders) > 0 && len(ob.SellOrders) > 0 {
		buyOrder := ob.BuyOrders[0]
		sellOrder := ob.SellOrders[0]
		if buyOrder.Price >= sellOrder.Price {
			tradeAmount := min(buyOrder.Amount, sellOrder.Amount)
			// 这里可以记录交易信息，如交易价格（sellOrder.Price）、交易数量tradeAmount等
			// 更新订单数量
			ob.BuyOrders[0].Amount -= tradeAmount
			ob.SellOrders[0].Amount -= tradeAmount
			if ob.BuyOrders[0].Amount == 0 {
				ob.BuyOrders = ob.BuyOrders[1:]
			}
			if ob.SellOrders[0].Amount == 0 {
				ob.SellOrders = ob.SellOrders[1:]
			}
		} else {
			break
		}
	}
	if len(ob.BuyOrders) == 0 && len(ob.SellOrders) == 0 {
		fmt.Println("All orders have been matched")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	// response := Response{
	// 	Code:    int(StatusOK),
	// 	Message: "Winning numbers generated successfully",
	// 	Data:    map[string]interface{}{"winningNumbers": winningNumbers},
	// }

	
	
	// Prepare the response
	response := requestPool.Get().(*Response)
	response.Code = int(StatusOK)
	response.Message = "Winning numbers generated successfully"
	response.Data = map[string]interface{}{"winningNumbers": winningNumbers}

	defer requestPool.Put(response)
	
		

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

// 新增好友接口
func handleAddFriend(c *gin.Context) {
	userId := c.Query("userId")
	friendId := c.Query("friendId")
	if userId == "" || friendId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing userId or friendId"})
		return
	}

	// 添加好友
	friends[userId] = append(friends[userId], Friend{ID: friendId, Name: "Friend Name"}) // 这里可以根据需要设置好友名称

	c.JSON(http.StatusOK, gin.H{"message": "Friend added successfully"})
}

// 刪除好友接口
func handleRemoveFriend(c *gin.Context) {
	userId := c.Query("userId")
	friendId := c.Query("friendId")
	if userId == "" || friendId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing userId or friendId"})
		return
	}

	// 刪除好友
	for i, friend := range friends[userId] {
		if friend.ID == friendId {
			friends[userId] = append(friends[userId][:i], friends[userId][i+1:]...) // 删除好友
			c.JSON(http.StatusOK, gin.H{"message": "Friend removed successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Friend not found"})
}

func handleAddOrder(c *gin.Context) {
	var order Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order data"})
		return
	}

	// Create OrderBook instance if not exists
	orderBook := &OrderBook{}

	if err := orderBook.AddOrder(order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Match orders after adding new order
	orderBook.MatchOrders()

	response := Response{
		Code:    int(StatusOK),
		Message: string(MsgSuccess),
		Data:    map[string]interface{}{"order": order},
	}
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
		api.POST("/add-friend", handleAddFriend)
		api.DELETE("/remove-friend", handleRemoveFriend)
		api.POST("/order", handleAddOrder)
	}

	log.Println("Server starting at :8888")
	if err := r.Run(":8888"); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	orderBook := OrderBook{}
	newOrderChan := make(chan Order)
	go func() {
		for order := range newOrderChan {
			orderBook.AddOrder(order)
			orderBook.MatchOrders()
		}
	}()
}
