package handlers

import (
	"PropertyAppBackend/config"
	database "PropertyAppBackend/db"
	"PropertyAppBackend/models"
	"PropertyAppBackend/services"
	"PropertyAppBackend/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
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
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
	User         models.User `json:"user"`
	Role         string `json:"role"`
}

// Send OTP
func SendOTP() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SendOTPRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.PhoneNumber == "" {
			logrus.Warn("Invalid OTP request payload")
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		cfg := config.GetCachedConfig()
		database.GetCachedClient()
		userCollection := database.GetUserCollection()
		otpCollection := database.GetOTPCollection()

		var existingUser models.User
		err = userCollection.FindOne(database.Ctx, bson.M{"phoneNumber": req.PhoneNumber}).Decode(&existingUser)
		isExistingUser := (err == nil)

    //User is trying to SignUp (Phone Number + Name Provided)
        if req.Name != "" {
            if isExistingUser {
                http.Error(w, "Phone number is already registered", http.StatusConflict)
                return
            }
        }

		    // User is trying to Login (Only Phone Number Provided)
        if req.Name == "" {
            if !isExistingUser {
                http.Error(w, "Phone number is not registered", http.StatusUnauthorized)
                return
            }
        }

		// Generate OTP
		otp, err := utils.GenerateOtp(req.PhoneNumber)
		if err != nil {
			logrus.Error("Failed to generate OTP:", err)
			http.Error(w, "Failed to generate OTP: "+err.Error(), http.StatusInternalServerError)
			return
		}

		//Store OTP in DB with Expiry
		expiresAt := time.Now().Add(time.Duration(cfg.OTPLifetimeMinutes) * time.Minute)
		otpRecord := models.OTPRecord{
			PhoneNumber: req.PhoneNumber,
			OTP:         otp,
			ExpiresAt:   expiresAt,
			CreatedAt:   time.Now(),
		}

		_, err = otpCollection.UpdateOne(database.Ctx, bson.M{"phoneNumber": req.PhoneNumber}, bson.M{"$set": otpRecord}, options.Update().SetUpsert(true))
		if err != nil {
			logrus.Error("Failed to store OTP in database:", err)
			http.Error(w, "Failed to store OTP: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// **Send OTP via SMS**
		messageBody := fmt.Sprintf("Your OTP is: %s. Valid for %d minutes.", otp, cfg.OTPLifetimeMinutes)
		err = services.SendSMS(req.PhoneNumber, messageBody)
		if err != nil {
			logrus.Error("Failed to send OTP via SMS:", err)
			http.Error(w, "Failed to send OTP via SMS: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// **Response**
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "OTP sent successfully",
			"isNewUserFlow": !isExistingUser,
		})
	}
}


// VerifyOTP handles verifying the OTP for a phone number
func VerifyOTP() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req VerifyOTPRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.PhoneNumber == "" || req.OTP == "" {
			logrus.Warn("Invalid VerifyOTP request payload")
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		cfg := config.GetCachedConfig()
		database.GetCachedClient()
		otpCollection := database.GetOTPCollection()
		userCollection := database.GetUserCollection()
		refreshTokenCollection := database.GetRefreshTokenCollection() // Get refresh token collection

		var storedOTP models.OTPRecord
		err = otpCollection.FindOne(database.Ctx, bson.M{"phoneNumber": req.PhoneNumber}).Decode(&storedOTP)
		if err != nil || time.Now().After(storedOTP.ExpiresAt) {
			otpCollection.DeleteOne(database.Ctx, bson.M{"phoneNumber": req.PhoneNumber})
			http.Error(w, "OTP expired or invalid", http.StatusUnauthorized)
			return
		}

		if storedOTP.OTP != req.OTP {
			logrus.Warn("Invalid OTP provided")
			http.Error(w, "Invalid OTP", http.StatusUnauthorized)
			return
		}
		otpCollection.DeleteOne(database.Ctx, bson.M{"phoneNumber": req.PhoneNumber})
		var user models.User
		err = userCollection.FindOne(database.Ctx, bson.M{"phoneNumber": req.PhoneNumber}).Decode(&user)
		if err == mongo.ErrNoDocuments {
			if req.Name == "" {
				http.Error(w, "Name required for new user registration", http.StatusBadRequest)
				return
			}
			newUser := models.User{
				Name:        req.Name,
				PhoneNumber: &req.PhoneNumber,
				Role:        models.RegularUser,
				CreatedBy:   user.ID,
			}
			insertResult, _ := userCollection.InsertOne(database.Ctx, newUser)
			userCollection.FindOne(database.Ctx, bson.M{"_id": insertResult.InsertedID}).Decode(&user)
		}

		accessToken, _ := utils.GenerateAccessToken(user.ID, cfg.JWTSecret)
		refreshToken, _ := utils.GenerateRefreshToken(user.ID, cfg.RefreshTokenSecret, cfg.RefreshTokenLifetimeHours)
		refreshTokenCollection.InsertOne(database.Ctx, models.RefreshToken{Token: refreshToken, UserID: user.ID, ExpiresAt: time.Now().Add(time.Duration(cfg.RefreshTokenLifetimeHours) * time.Hour)})
		json.NewEncoder(w).Encode(AuthResponse{
			Message:      "User successfully verified",
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         user,
			Role:         string(user.Role),
		})
	}
}

// Mini-Admin Login API
func MiniAdminLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.User
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.Username == nil || req.Password == nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// **Find Mini-Admin**
		var miniAdmin models.User
		err = database.GetUserCollection().FindOne(r.Context(), bson.M{"username": req.Username, "role": models.MiniAdmin}).Decode(&miniAdmin)
		if err != nil {
			http.Error(w, "Mini-Admin not found", http.StatusUnauthorized)
			return
		}

		// **Check Password**
		if !utils.CheckPasswordHash(*req.Password, *miniAdmin.Password) {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}

		// **Generate Token**
		token, _ := utils.GenerateJWT(miniAdmin.ID, string(models.MiniAdmin))

		// **Return Response**
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "Login successful",
			"userID":   miniAdmin.ID.Hex(),
			"username": miniAdmin.Username,
			"token":    token,
			"role":     miniAdmin.Role,
		})
	}
}
