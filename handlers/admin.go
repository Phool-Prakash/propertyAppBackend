package handlers

import (
	database "PropertyAppBackend/db"
	"PropertyAppBackend/models"
	"PropertyAppBackend/utils"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// Admin Login API
func AdminLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.User
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.Username == nil || req.Password == nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// **Find Admin**
		var admin models.User
		err = database.GetUserCollection().FindOne(r.Context(), bson.M{"username": req.Username, "role": models.Admin}).Decode(&admin)
		if err != nil {
			http.Error(w, "Admin not found", http.StatusUnauthorized)
			return
		}

		// **Check Password**
		if !utils.CheckPasswordHash(*req.Password, *admin.Password) {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}

		// **Generate Token**
		token, _ := utils.GenerateJWT(admin.ID, string(admin.Role))

		// **Return Response**
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "Login successful",
			"userID":   admin.ID.Hex(),
			"username": admin.Username,
			"token":    token,
		})
	}
}

// create Mini Admin

func CreateMiniAdmin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Token is missing", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		println("TokenString", tokenString)
		adminID, claims, err := utils.ParseJWT(tokenString, string(utils.JWTSecret))
		if err != nil {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			println(err)
			return
		}
		role, ok := claims["role"].(string)
		if !ok || role != string(models.Admin) {
			http.Error(w, "Unauthorized: Only Admin can perform this action", http.StatusForbidden)
			return
		}
		var req models.User
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.Username ==  nil || req.Password == nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		//Check if the request is from admin
		adminCheck, _ := database.GetUserCollection().CountDocuments(r.Context(), bson.M{"role": models.Admin})
		if adminCheck == 0 {
			http.Error(w, "Only Admin can create Mini-Admin!", http.StatusForbidden)
			return
		}

		existingCount, _ := database.GetUserCollection().CountDocuments(r.Context(), bson.M{"username": req.Username, "role": models.MiniAdmin})
		if existingCount > 0 {
			http.Error(w, "Mini-Admin with this username already exists", http.StatusConflict)
			return
		}

		//hashPassword

		hashedPassword, _ := utils.HashPassword(*req.Password)
		if err != nil {
    http.Error(w, "Failed to hash password", http.StatusInternalServerError)
    return
}
		newMiniAdmin := models.User{
			Username:  req.Username,
			Password:  &hashedPassword,
			Role:      models.MiniAdmin,
			CreatedBy: adminID,
			CreatedAt: time.Now(),
		}

		_, err = database.GetUserCollection().InsertOne(r.Context(), newMiniAdmin)
		if err != nil {
			http.Error(w, "Failed to create Mini-Admin", http.StatusInternalServerError)
			 logrus.WithError(err).Error("Failed to create Mini-Admin")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Mini-Admin created successfully!",
		"username": *req.Username,
		"password": *req.Password,
		"role": string(req.Role),
		})
	}
}
