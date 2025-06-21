package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string

const (
    Admin     Role = "admin"
    MiniAdmin Role = "mini-admin"
    RegularUser Role = "user"
)

type User struct {
    ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    Name        string             `json:"name" bson:"name"`
    PhoneNumber *string             `json:"phoneNumber" bson:"phoneNumber,omitempty" validate:"min=10,max=13"`
    Username    *string             `json:"username,omitempty" bson:"username,omitempty" validate:"min=5,max=20"`
    Password *string `json:"password,omitempty" bson:"password,omitempty" validate:"min=6,max=20"`
    Role        Role               `json:"role" bson:"role,omitempty"`
    CreatedAt   time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	CreatedBy  primitive.ObjectID `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
}

//OTPRecord
type OTPRecord struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	PhoneNumber string             `json:"phoneNumber" bson:"phoneNumber"`
	OTP         string             `json:"otp" bson:"otp"`
	ExpiresAt   time.Time          `json:"expiresAt" bson:"expiresAt"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
}

// ... existing code ...

// type MiniAdmin struct {
//     ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
//     Username  string             `json:"username" bson:"username" validate:"required,min=5,max=20"`
//     Password  string             `json:"password" bson:"password" validate:"required,min=6,max=20"`
//     Role      Role               `json:"role" bson:"role"`
//     CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
//     CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
// }