package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type      string             `bson:"type" json:"type"`
	Sender    string             `bson:"sender" json:"sender"`
	Receiver  string             `bson:"receiver" json:"receiver"`
	Message   string             `bson:"message" json:"message"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}
