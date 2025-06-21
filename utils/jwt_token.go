package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var JWTSecret = []byte("Min-Admin@123")

// GenerateJWT generates a JWT token for Admin, Mini-Admin, or User
func GenerateJWT(userID primitive.ObjectID, role string) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID.Hex(),
        "role":    role,
        "exp":     jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(JWTSecret)
    if err != nil {
        return "", fmt.Errorf("failed to sign JWT token: %w", err)
    }
    return tokenString, nil
}
// GenerateAccessToken generates a new short-lived JWT Access Token
func GenerateAccessToken(userID primitive.ObjectID, secret string) (string, error) {
	claims := jwt.MapClaims{
		"authorized": true,
		"user_id":    userID.Hex(),
		"exp":        jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return tokenString, nil
}

// GenerateRefreshToken generates a new long-lived JWT Refresh Token
func GenerateRefreshToken(userID primitive.ObjectID, secret string, lifetimeHours int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.Hex(),
		"exp":     jwt.NewNumericDate(time.Now().Add(time.Duration(lifetimeHours) * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return tokenString, nil
}

// ParseJWT parses a JWT token (Access or Refresh) and returns user ID and claims
func ParseJWT(tokenString string, secret string) (primitive.ObjectID, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return primitive.NilObjectID, nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return primitive.NilObjectID, nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return primitive.NilObjectID, nil, fmt.Errorf("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return primitive.NilObjectID, nil, fmt.Errorf("user_id not found or invalid in token claims")
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return primitive.NilObjectID, nil, fmt.Errorf("invalid user_id format in token claims: %w", err)
	}

	return userID, claims, nil
}