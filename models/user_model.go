package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" `
	Name        string             `json:"name" bson:"name"`
	PhoneNumber string             `json:"phoneNumber" bson:"phoneNumber, omitempty" validate:"required,min=10,max=13"`
	CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

// type OTPRecord struct {
// 	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
// 	PhoneNumber string             `json:"phoneNumber" bson:"phoneNumber"`
// 	OTP         string             `json:"otp" bson:"otp"`
// 	ExpiresAt   time.Time          `json:"expiresAt" bson:"expiresAt"`
// 	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
// }