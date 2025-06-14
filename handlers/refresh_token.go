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
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshResponse is the response for a successful token refresh
type RefreshResponse struct {
	Message     string      `json:"message"`
	AccessToken string      `json:"accessToken"`
	User        models.User `json:"user"`
}

func RefreshAccessToken() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req RefreshRequest
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
        json.NewEncoder(w).Encode(RefreshResponse{
            Message:     "Tokens refreshed successfully",
            AccessToken: newAccessToken,
            User:        user,
        })
    }
}




// // RefreshAccessToken handles the refresh token request
// func RefreshAccessToken(client *mongo.Client) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var req RefreshRequest
// 		err := json.NewDecoder(r.Body).Decode(&req)
// 		if err != nil {
// 			http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 			return
// 		}

// 		if req.RefreshToken == "" {
// 			http.Error(w, "Refresh token is required", http.StatusBadRequest)
// 			return
// 		}

// 		cfg := config.LoadConfig()
// 		refreshTokenCollection := database.GetRefreshTokenCollection(client)
// 		userCollection := database.GetUserCollection(client)

// 		// 1. Parse and validate the refresh token
// 		userID, _, err := utils.ParseJWT(req.RefreshToken, cfg.RefreshTokenSecret)
// 		if err != nil {
// 			log.Printf("Refresh Token Error: Invalid refresh token provided: %v", err)
// 			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
// 			return
// 		}

// 		// 2. Check if the refresh token exists in the database and is not expired
// 		var storedRefreshToken models.RefreshToken
// 		err = refreshTokenCollection.FindOne(database.Ctx, bson.M{"token": req.RefreshToken, "userId": userID}).Decode(&storedRefreshToken)
// 		if err != nil {
// 			if err == mongo.ErrNoDocuments {
// 				log.Printf("Refresh Token Error: Refresh token not found in DB for user %s or already used.", userID.Hex())
// 				http.Error(w, "Refresh token not found or already used. Please login again.", http.StatusUnauthorized)
// 			} else {
// 				log.Printf("Refresh Token Error: DB error while looking up refresh token for user %s: %v", userID.Hex(), err)
// 				http.Error(w, "Database error", http.StatusInternalServerError)
// 			}
// 			return
// 		}

// 		// Check if stored refresh token has expired (although JWT parse should catch this)
// 		if time.Now().After(storedRefreshToken.ExpiresAt) {
// 			log.Printf("Refresh Token Error: Stored refresh token for user %s has expired.", userID.Hex())
// 			// Delete expired refresh token from DB
// 			refreshTokenCollection.DeleteOne(database.Ctx, bson.M{"_id": storedRefreshToken.ID})
// 			http.Error(w, "Refresh token has expired. Please login again.", http.StatusUnauthorized)
// 			return
// 		}

// 		// 3. Invalidate the old refresh token (Optional but Recommended for single-use refresh tokens)
// 		// This prevents replay attacks if the refresh token is stolen.
// 		_, err = refreshTokenCollection.DeleteOne(database.Ctx, bson.M{"_id": storedRefreshToken.ID})
// 		if err != nil {
// 			log.Printf("Refresh Token Warning: Failed to delete used refresh token %s for user %s: %v", req.RefreshToken, userID.Hex(), err)
// 			// Log but don't fail the request unless critical
// 		}

// 		// 4. Retrieve user information
// 		var user models.User
// 		err = userCollection.FindOne(database.Ctx, bson.M{"_id": userID}).Decode(&user)
// 		if err != nil {
// 			if err == mongo.ErrNoDocuments {
// 				log.Printf("Refresh Token Error: User not found for ID %s associated with refresh token.", userID.Hex())
// 				http.Error(w, "User not found", http.StatusNotFound)
// 			} else {
// 				log.Printf("Refresh Token Error: DB error while fetching user for ID %s: %v", userID.Hex(), err)
// 				http.Error(w, "Database error", http.StatusInternalServerError)
// 			}
// 			return
// 		}

// 		// 5. Generate a new Access Token
// 		newAccessToken, err := utils.GenerateAccessToken(user.ID, cfg.JWTSecret)
// 		if err != nil {
// 			log.Printf("Refresh Token Error: Failed to generate new access token for user %s: %v", user.ID.Hex(), err)
// 			http.Error(w, "Failed to generate new access token", http.StatusInternalServerError)
// 			return
// 		}

// 		// 6. Generate a new Refresh Token (Optional: Rotate Refresh Tokens)
// 		// This is a common security practice known as "Refresh Token Rotation".
// 		// It means every time a refresh token is used, a new one is issued,
// 		// and the old one is immediately invalidated.
// 		newRefreshToken, err := utils.GenerateRefreshToken(user.ID, cfg.RefreshTokenSecret, cfg.RefreshTokenLifetimeHours)
// 		if err != nil {
// 			log.Printf("Refresh Token Error: Failed to generate new refresh token for user %s: %v", user.ID.Hex(), err)
// 			http.Error(w, "Failed to generate new refresh token", http.StatusInternalServerError)
// 			return
// 		}

// 		// 7. Store the new refresh token in the database
// 		newRefreshTokenRecord := models.RefreshToken{
// 			Token:     newRefreshToken,
// 			UserID:    user.ID,
// 			ExpiresAt: time.Now().Add(time.Duration(cfg.RefreshTokenLifetimeHours) * time.Hour),
// 			CreatedAt: time.Now(),
// 		}
// 		_, err = refreshTokenCollection.InsertOne(database.Ctx, newRefreshTokenRecord)
// 		if err != nil {
// 			log.Printf("Refresh Token Warning: Failed to store new refresh token for user %s: %v", user.ID.Hex(), err)
// 			// Log and proceed, as Access Token is already generated.
// 		}

// 		w.WriteHeader(http.StatusOK)
// 		json.NewEncoder(w).Encode(RefreshResponse{
// 			Message:     "Tokens refreshed successfully",
// 			AccessToken: newAccessToken,
// 			User:        user, // Return user info
// 		})
// 	}
// }