// This file handle MongoDB connection and Database Collections
// To estiblish the database connection and return the specific collections
// package database

// import (
// 	"context"
// 	"log"
// 	"sync"

// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// var once sync.Once
// var cachedClient *mongo.Client
// var Ctx = context.Background()

// func ConnectDB(uri string) (*mongo.Client, error) {
// 	var err error
// 	once.Do(func ()  {
// 		clientOptions := options.Client().ApplyURI(uri).SetMaxPoolSize(50)
// 		_, err := mongo.Connect(Ctx,clientOptions)
// 	if err != nil {
// 		log.Fatal("Database connection failed:",err)
// 	} 
// 	})
// 	return cachedClient, err
// }

// // GetUserCollection returns the users collection
// func GetUserCollection(client *mongo.Client) *mongo.Collection {
// 	return client.Database("propertyAppDatabase").Collection("users")
// }

// // GetOTPCollection returns the otps collection
// func GetOTPCollection(client *mongo.Client) *mongo.Collection {
// 	return client.Database("propertyAppDatabase").Collection("otps")
// }

// func GetRefreshTokenCollection(client *mongo.Client) *mongo.Collection { // <-- NEW
// 	return client.Database("propertyAppDatabase").Collection("refresh_tokens")
// }



package database

import (
    "context"
    "log"
    "sync"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var once sync.Once
var cachedClient *mongo.Client
var Ctx = context.Background()

// ConnectDB with caching and retry mechanism
func ConnectDB(uri string) (*mongo.Client, error) {
    var err error
    once.Do(func() {
        clientOptions := options.Client().ApplyURI(uri).SetMaxPoolSize(50)

        cachedClient, err = mongo.Connect(Ctx, clientOptions)
        if err != nil {
            log.Fatal("Database connection failed:", err)
        }
    })

    return cachedClient, err
}

// Get specific collections

//GetUserCollection returns the users collection
func GetUserCollection(client *mongo.Client) *mongo.Collection {
	if cachedClient == nil {
		log.Println("Database client not initialized!")
	}
    return client.Database("propertyAppDatabase").Collection("users")
}

func GetOTPCollection(client *mongo.Client) *mongo.Collection {
	if cachedClient == nil {
		log.Println("Database client not initialized!")

	}
    return client.Database("propertyAppDatabase").Collection("otps")
}

func GetRefreshTokenCollection(client *mongo.Client) *mongo.Collection {
	if cachedClient == nil {
		log.Println("Database client not initialized!")

	}
    return client.Database("propertyAppDatabase").Collection("refresh_tokens")
}
