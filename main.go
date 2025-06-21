package main

import (
	"PropertyAppBackend/config"
	database "PropertyAppBackend/db"
	"PropertyAppBackend/handlers"
	"PropertyAppBackend/middleware"
	"PropertyAppBackend/utils"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func main() {
	hashedPassword, _ := utils.HashPassword("DefaultAdmin@123")
	 fmt.Println("Admin Hashed Password:", hashedPassword)
	config.LoadConfig()
	// Load configuration
	cfg := config.GetCachedConfig()
	// Connect to MongoDB once at startup
	client, err := database.ConnectDB(cfg.MongoDBURI)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err) // Fatal if connection fails
	}
	defer func() {
		// Proper disconnection when app closes
		if err = client.Disconnect(context.Background()); err != nil { // Use context.Background() for defer
			log.Printf("Error disconnecting from MongoDB: %v", err)
		} else {
			fmt.Println("Disconnected from MongoDB.")
		}
	}()
	fmt.Println("Connected to MongoDB!")

	database.SeedAdminUser()

	// Ensure unique index on phoneNumber for users collection
	userCollection := database.GetUserCollection()
indexes := []mongo.IndexModel{
    {Keys: bson.D{{Key: "phoneNumber", Value: 1}}, Options: options.Index().SetUnique(true)}, // ✅ Unique Phone for Users
    {Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},   // ✅ Unique Username for Admin & Mini-Admin
}
	_, err = userCollection.Indexes().CreateMany(database.Ctx, indexes)
	if err != nil {
		// Log error if index creation fails, could be due to existing duplicate data
		log.Printf("Warning: Failed to create unique index for users collection, or index already exists: %v", err)
	} else {
		fmt.Println("Unique indexes ensured: phoneNumber (Users) & username (Admins)")
	}

	r := mux.NewRouter()

	// Authentication routes - Pass the MongoDB client to handlers
	r.HandleFunc("/send-otp", handlers.SendOTP()).Methods("POST")
	r.HandleFunc("/verify-otp", handlers.VerifyOTP()).Methods("POST")


// **Admin Creates Mini-Admin**
r.HandleFunc("/admin/create-mini-admin", handlers.CreateMiniAdmin()).Methods("POST")

// **Admin Login**
r.HandleFunc("/admin/login", handlers.AdminLogin()).Methods("POST")
r.HandleFunc("/mini-admin/login", handlers.MiniAdminLogin()).Methods("POST")



	// Protected routes (require authentication via JWT)
	protectedRouter := r.PathPrefix("/api").Subrouter()
	// Pass client to middleware if middleware needs DB access, else no change
	protectedRouter.Use(middleware.AuthMiddleware) // No client needed for basic AuthMiddleware
	// protectedRouter.HandleFunc("/profile", handlers.GetUserProfile(client)).Methods("GET")

	fmt.Printf("Server listening on %s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, r))
}


