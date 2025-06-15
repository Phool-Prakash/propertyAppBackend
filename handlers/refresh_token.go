package handlers

import (
	"PropertyAppBackend/config"
	database "PropertyAppBackend/db"
	"PropertyAppBackend/models"
	"PropertyAppBackend/utils"
	"encoding/json"
	"net/http"
	"time"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// RefreshRequest is the payload for refreshing tokens


func RefreshAccessToken() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req models.RefreshRequest
        err := json.NewDecoder(r.Body).Decode(&req)
        if err != nil || req.RefreshToken == "" {
			logrus.Warn("Invalid refresh token request payload")
            http.Error(w, "Invalid request payload", http.StatusBadRequest)
            return
        }

        // Corrected: Retrieve Cached Config & Database Client
        cfg:= config.GetCachedConfig()
         database.GetCachedClient()

        // Corrected: Pass client as an argument
        refreshTokenCollection := database.GetRefreshTokenCollection()
        userCollection := database.GetUserCollection()  

        // Validate Refresh Token
        userID, _, err := utils.ParseJWT(req.RefreshToken, cfg.RefreshTokenSecret)
        if err != nil {
			logrus.Warn("Invalid refresh token provided")
            http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
            return
        }

        // Check Refresh Token Existence & Expiry
        var storedRefreshToken models.RefreshToken
        err = refreshTokenCollection.FindOne(r.Context(), bson.M{"token": req.RefreshToken, "userId": userID}).Decode(&storedRefreshToken)
        if err != nil || time.Now().After(storedRefreshToken.ExpiresAt) {
			logrus.Warn("Refresh token expired or not found")
            refreshTokenCollection.DeleteOne(r.Context(), bson.M{"_id": storedRefreshToken.ID})
            http.Error(w, "Refresh token expired or invalid, login required.", http.StatusUnauthorized)
            return
        }

        // Invalidate Old Refresh Token (Security Best Practice)
        _, _ = refreshTokenCollection.DeleteOne(r.Context(), bson.M{"_id": storedRefreshToken.ID})

        // Fetch User Data
        var user models.User
        err = userCollection.FindOne(r.Context(), bson.M{"_id": userID}).Decode(&user)
        if err != nil {
			logrus.Warn("User not found for refresh token")
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }

        // Generate New Access & Refresh Token
        newAccessToken, _ := utils.GenerateAccessToken(user.ID, cfg.JWTSecret)
        newRefreshToken, _ := utils.GenerateRefreshToken(user.ID, cfg.RefreshTokenSecret, cfg.RefreshTokenLifetimeHours)

        // Store New Refresh Token
        newRefreshTokenRecord := models.RefreshToken{
            Token:     newRefreshToken,
            UserID:    user.ID,
            ExpiresAt: time.Now().Add(time.Duration(cfg.RefreshTokenLifetimeHours) * time.Hour),
            CreatedAt: time.Now(),
        }
        _, _ = refreshTokenCollection.InsertOne(r.Context(), newRefreshTokenRecord)

		logrus.Info("Tokens refreshed successfully for user:", user.ID.Hex())

        // Send Response

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(models.RefreshResponse{
            Message:     "Tokens refreshed successfully",
            AccessToken: newAccessToken,
            User:        user,
        })
    }
}
