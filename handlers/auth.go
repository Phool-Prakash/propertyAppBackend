package handlers

import (
	"PropertyAppBackend/config"
	database "PropertyAppBackend/db"
	"PropertyAppBackend/models"
	"PropertyAppBackend/services"
	"PropertyAppBackend/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Request/Response Structs (Remains same)
type SendOTPRequest struct {
	Name        string `json:"name,omitempty"`
	PhoneNumber string `json:"phoneNumber"`
}

type VerifyOTPRequest struct {
	Name        string `json:"name,omitempty"`
	PhoneNumber string `json:"phoneNumber"`
	OTP         string `json:"otp"`
}

// AuthResponse will now include RefreshToken
type AuthResponse struct {
	Message      string      `json:"message"`
	AccessToken  string      `json:"accessToken"`  // Changed to AccessToken
	RefreshToken string      `json:"refreshToken"` // NEW
	User         models.User `json:"user"`
}

// --- API Handlers ---

// SendOTP handles sending an OTP to a phone number
func SendOTP(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SendOTPRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if req.PhoneNumber == "" {
			http.Error(w, "Phone number is required", http.StatusBadRequest)
			return
		}

		cfg := config.LoadConfig()
		userCollection := database.GetUserCollection()

		var existingUser models.User
		err = userCollection.FindOne(database.Ctx, bson.M{"phoneNumber": req.PhoneNumber}).Decode(&existingUser)
		isExistingUser := (err == nil)

		if req.Name != "" && isExistingUser && existingUser.Name != req.Name {
			http.Error(w, "User with this phone number already exists with a different name.", http.StatusConflict)
			return
		}

		if req.Name == "" && !isExistingUser {
			http.Error(w, "Name is required for new user signup. If logging in, do not provide name in send-otp.", http.StatusBadRequest)
			return
		}

		otp, err := utils.GenerateOtp(req.PhoneNumber)
		if err != nil {
			http.Error(w, "Failed to generate OTP: "+err.Error(), http.StatusInternalServerError)
			return
		}

		otpCollection := database.GetOTPCollection()

		expiresAt := time.Now().Add(time.Duration(cfg.OTPLifetimeMinutes) * time.Minute)
		otpRecord := models.OTPRecord{
			PhoneNumber: req.PhoneNumber,
			OTP:         otp,
			ExpiresAt:   expiresAt,
			CreatedAt:   time.Now(),
		}

		filter := bson.M{"phoneNumber": req.PhoneNumber}
		update := bson.M{"$set": otpRecord}
		opts := options.Update().SetUpsert(true)

		_, err = otpCollection.UpdateOne(database.Ctx, filter, update, opts)
		if err != nil {
			log.Printf("ERROR: Failed to store OTP for %s: %v", req.PhoneNumber, err)
			http.Error(w, "Failed to store OTP: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("OTP for %s stored successfully in DB.", req.PhoneNumber)

		messageBody := fmt.Sprintf("Your OTP for verification is: %s. It is valid for %d minutes.", otp, cfg.OTPLifetimeMinutes)
		err = services.SendSMS(req.PhoneNumber, messageBody)
		if err != nil {
			log.Printf("ERROR: Failed to send OTP via SMS for %s: %v", req.PhoneNumber, err)
			http.Error(w, "Failed to send OTP via SMS: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("OTP SMS sent to %s.", req.PhoneNumber)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "OTP sent successfully",
			"isNewUserFlow": !isExistingUser,
		})
	}
}

// VerifyOTP handles verifying the OTP for a phone number
// This API also handles user registration/login based on OTP verification result.
func VerifyOTP(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req VerifyOTPRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if req.PhoneNumber == "" || req.OTP == "" {
			http.Error(w, "Phone number and OTP are required", http.StatusBadRequest)
			return
		}

		cfg := config.LoadConfig()
		otpCollection := database.GetOTPCollection()
		userCollection := database.GetUserCollection()
		refreshTokenCollection := database.GetRefreshTokenCollection() // Get refresh token collection

		var storedOTP models.OTPRecord
		err = otpCollection.FindOne(database.Ctx, bson.M{"phoneNumber": req.PhoneNumber}).Decode(&storedOTP)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				log.Printf("VerifyOTP: No OTP record found for %s. Error: %v", req.PhoneNumber, err)
				http.Error(w, "No OTP found for this phone number. Please request a new one.", http.StatusBadRequest)
			} else {
				log.Printf("VerifyOTP: Database error finding OTP for %s: %v", req.PhoneNumber, err)
				http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}
		log.Printf("VerifyOTP: Found OTP record for %s. Stored OTP: %s, ExpiresAt: %v", req.PhoneNumber, storedOTP.OTP, storedOTP.ExpiresAt)

		if time.Now().After(storedOTP.ExpiresAt) {
			log.Printf("VerifyOTP: OTP for %s expired. ExpiresAt: %v, Current: %v", req.PhoneNumber, storedOTP.ExpiresAt, time.Now())
			otpCollection.DeleteOne(database.Ctx, bson.M{"phoneNumber": req.PhoneNumber})
			http.Error(w, "OTP has expired. Please request a new one.", http.StatusUnauthorized)
			return
		}
		log.Printf("VerifyOTP: OTP for %s is not expired.", req.PhoneNumber)

		if storedOTP.OTP != req.OTP {
			log.Printf("VerifyOTP: Invalid OTP provided for %s. Expected: %s, Got: %s", req.PhoneNumber, storedOTP.OTP, req.OTP)
			http.Error(w, "Invalid OTP.", http.StatusUnauthorized)
			return
		}
		log.Printf("VerifyOTP: OTP for %s matched.", req.PhoneNumber)

		_, err = otpCollection.DeleteOne(database.Ctx, bson.M{"phoneNumber": req.PhoneNumber})
		if err != nil {
			fmt.Printf("Warning: Failed to delete OTP for %s after verification: %v\n", req.PhoneNumber, err)
		} else {
			log.Printf("OTP for %s deleted from DB after verification.", req.PhoneNumber)
		}

		var user models.User
		err = userCollection.FindOne(database.Ctx, bson.M{"phoneNumber": req.PhoneNumber}).Decode(&user)

		if err == mongo.ErrNoDocuments {
			if req.Name == "" {
				log.Printf("VerifyOTP: New user signup attempt for %s without name.", req.PhoneNumber)
				http.Error(w, "Name is required to register a new user.", http.StatusBadRequest)
				return
			}

			newUser := models.User{
				Name:        req.Name,
				PhoneNumber: req.PhoneNumber,
			}

			insertResult, err := userCollection.InsertOne(database.Ctx, newUser)
			if err != nil {
				if mongo.IsDuplicateKeyError(err) {
					log.Printf("VerifyOTP: Duplicate key error during new user registration for %s.", req.PhoneNumber)
					http.Error(w, "User with this phone number already registered. Please login.", http.StatusConflict)
					return
				}
				log.Printf("VerifyOTP: Error registering new user %s: %v", req.PhoneNumber, err)
				http.Error(w, "Error registering new user: "+err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("VerifyOTP: New user %s registered with ID: %s", req.PhoneNumber, insertResult.InsertedID)

			err = userCollection.FindOne(database.Ctx, bson.M{"_id": insertResult.InsertedID}).Decode(&user)
			if err != nil {
				log.Printf("VerifyOTP: Failed to retrieve new user %s after registration: %v", req.PhoneNumber, err)
				http.Error(w, "Failed to retrieve new user after registration: "+err.Error(), http.StatusInternalServerError)
				return
			}

		} else if err != nil {
			log.Printf("VerifyOTP: Database error during user lookup for %s: %v", req.PhoneNumber, err)
			http.Error(w, "Database error during user lookup: "+err.Error(), http.StatusInternalServerError)
			return
		} else {
			if req.Name != "" && user.Name != req.Name {
				log.Printf("VerifyOTP: Name mismatch for existing user %s. Provided: %s, Stored: %s", req.PhoneNumber, req.Name, user.Name)
				http.Error(w, "Provided name does not match existing user's name for this phone number.", http.StatusConflict)
				return
			}
			log.Printf("VerifyOTP: Existing user %s logged in.", req.PhoneNumber)
		}

		// --- NEW: Generate Access Token and Refresh Token ---
		accessToken, err := utils.GenerateAccessToken(user.ID, cfg.JWTSecret)
		if err != nil {
			log.Printf("VerifyOTP: Error generating Access Token for user %s: %v", user.ID.Hex(), err)
			http.Error(w, "Error generating access token", http.StatusInternalServerError)
			return
		}

		refreshToken, err := utils.GenerateRefreshToken(user.ID, cfg.RefreshTokenSecret, cfg.RefreshTokenLifetimeHours)
		if err != nil {
			log.Printf("VerifyOTP: Error generating Refresh Token for user %s: %v", user.ID.Hex(), err)
			http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
			return
		}

		// --- NEW: Store Refresh Token in database ---
		refreshTokenRecord := models.RefreshToken{
			Token:     refreshToken,
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(time.Duration(cfg.RefreshTokenLifetimeHours) * time.Hour),
			CreatedAt: time.Now(),
		}
		_, err = refreshTokenCollection.InsertOne(database.Ctx, refreshTokenRecord)
		if err != nil {
			log.Printf("VerifyOTP: Failed to store refresh token for user %s: %v", user.ID.Hex(), err)
			// This is a critical error, but we might still return tokens.
			// Depending on strictness, you might return 500 here.
			// For now, logging and continuing.
		}
		log.Printf("Refresh Token for user %s stored successfully in DB.", user.ID.Hex())

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(AuthResponse{
			Message:      "User successfully verified",
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         user,
		})
	}
}

// GetUserProfile is an example of a protected route
func GetUserProfile(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Context().Value("userID")
		if userIDStr == nil {
			http.Error(w, "User ID not found in context.", http.StatusInternalServerError)
			return
		}
		userID, ok := userIDStr.(primitive.ObjectID) // Assuming middleware puts ObjectID, not string
		if !ok {
			http.Error(w, "Invalid User ID type in context.", http.StatusInternalServerError)
			return
		}

		config.LoadConfig()
		userCollection := database.GetUserCollection()

		var user models.User
		err := userCollection.FindOne(database.Ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		json.NewEncoder(w).Encode(user)
	}
}
