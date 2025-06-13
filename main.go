package main

import (
	"PropertyAppBackend/config"
	database "PropertyAppBackend/db"
	"PropertyAppBackend/handlers"
	"PropertyAppBackend/middleware"
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
	// Load configuration
	cfg := config.LoadConfig()

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

	// Ensure unique index on phoneNumber for users collection
	userCollection := database.GetUserCollection(client)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"phoneNumber", 1}},       // Index on phoneNumber field
		Options: options.Index().SetUnique(true),  // Make it unique
	}
	_, err = userCollection.Indexes().CreateOne(database.Ctx, indexModel)
	if err != nil {
		// Log error if index creation fails, could be due to existing duplicate data
		log.Printf("Warning: Failed to create unique index for users collection, or index already exists: %v", err)
	} else {
		fmt.Println("Unique index ensured on users.phoneNumber")
	}

	r := mux.NewRouter()

	// Authentication routes - Pass the MongoDB client to handlers
	r.HandleFunc("/send-otp", handlers.SendOTP(client)).Methods("POST")
	r.HandleFunc("/verify-otp", handlers.VerifyOTP(client)).Methods("POST")

	// Protected routes (require authentication via JWT)
	protectedRouter := r.PathPrefix("/api").Subrouter()
	// Pass client to middleware if middleware needs DB access, else no change
	protectedRouter.Use(middleware.AuthMiddleware) // No client needed for basic AuthMiddleware
	// protectedRouter.HandleFunc("/profile", handlers.GetUserProfile(client)).Methods("GET")

	fmt.Printf("Server listening on %s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, r))
}


// func main() {
// 	// Load configuration
// 	cfg := config.LoadConfig()

// 	// Connect to MongoDB
// 	client, err := database.ConnectDB(cfg.MongoDBURI)
// 	if err != nil {
// 		log.Fatalf("MongoDB connection error: %v", err)
// 	}
// 	defer func() {
// 		if err = client.Disconnect(database.Ctx); err != nil {
// 			log.Fatal(err)
// 		}
// 	}()
// 	fmt.Println("Connected to MongoDB!")

// 	// Ensure unique index on phoneNumber for users collection
// 	userCollection := database.GetUserCollection(client)
// 	indexModel := mongo.IndexModel{
// 		Keys:    bson.D{{"phoneNumber", 1}},      // Index on phoneNumber field
// 		Options: options.Index().SetUnique(true), // Make it unique
// 	}
// 	_, err = userCollection.Indexes().CreateOne(database.Ctx, indexModel)
// 	if err != nil {
// 		// Log error if index creation fails, could be due to existing duplicate data
// 		log.Printf("Warning: Failed to create unique index for users collection, or index already exists: %v", err)
// 	} else {
// 		fmt.Println("Unique index ensured on users.phoneNumber")
// 	}

// 	r := mux.NewRouter()

// 	// Authentication routes
// 	r.HandleFunc("/send-otp", handlers.SendOTP).Methods("POST")
// 	r.HandleFunc("/verify-otp", handlers.VerifyOTP).Methods("POST")

// 	// Protected routes (require authentication via JWT)
// 	protectedRouter := r.PathPrefix("/api").Subrouter()
// 	protectedRouter.Use(middleware.AuthMiddleware)
// 	protectedRouter.HandleFunc("/profile", handlers.GetUserProfile).Methods("GET")

// 	fmt.Printf("Server listening on %s\n", cfg.Port)
// 	log.Fatal(http.ListenAndServe(cfg.Port, r))
// }
