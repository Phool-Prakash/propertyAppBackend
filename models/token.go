package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshResponse is the response for a successful token refresh
type RefreshResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"accessToken"`
	User        User   `json:"user"`
}

type RefreshToken struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Token     string             `json:"token" bson:"token"` // The JWT string for the refresh token
	UserID    primitive.ObjectID `json:"userId" bson:"userId"`
	ExpiresAt time.Time          `json:"expiresAt" bson:"expiresAt"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	// You can add more fields like:
	// IPAddress string `json:"ipAddress" bson:"ipAddress"`
	// UserAgent string `json:"userAgent" bson:"userAgent"`
	// IsRevoked bool   `json:"isRevoked" bson:"isRevoked"` // For manual revocation
}
