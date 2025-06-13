package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OTPRecord represents an OTP stored in the database for verification
type OTPRecord struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	PhoneNumber string             `json:"phoneNumber" bson:"phoneNumber"`
	OTP         string             `json:"otp" bson:"otp"`
	ExpiresAt   time.Time          `json:"expiresAt" bson:"expiresAt"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
}