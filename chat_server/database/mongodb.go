package database

import (
	"context"
	"log"
	"time"

	"p2p_chat/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client *mongo.Client
	ChatDB *mongo.Database
)

func ConnectMongoDB(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// 測試連接
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	Client = client
	ChatDB = client.Database(cfg.DBName)

	log.Printf("Successfully connected to MongoDB at %s", cfg.MongoURI)
	return nil
}

func CloseMongoDB() {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := Client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}
}
