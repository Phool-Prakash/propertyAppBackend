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
	"fmt"
	"log"
	"sync"
	"time"

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
        clientOptions := options.Client().ApplyURI(uri).SetMaxPoolSize(50).SetConnectTimeout(10 * time.Second)

        cachedClient, err = mongo.Connect(Ctx, clientOptions)
        if err != nil {
            log.Println("Database connection failed:", err)
			err = fmt.Errorf("Database connection failed: %w", err)
        }
    })

    return cachedClient, err
}

// Retrieve shared DB client instance
func GetCachedClient() *mongo.Client {
    return cachedClient
}

// Get specific collections

//GetUserCollection returns the users collection
func GetUserCollection() *mongo.Collection {
	if cachedClient == nil {
		log.Println("Database client not initialized!")
		return nil
	}
    return cachedClient.Database("propertyAppDatabase").Collection("users")
}

//GetOTPCollection returns the OTPs collection
func GetOTPCollection() *mongo.Collection {
	if cachedClient == nil {
		log.Println("Database client not initialized!")
		return nil

	}
    return cachedClient.Database("propertyAppDatabase").Collection("otps")
}

//GetRefreshTokenCollection returns the refresh token collection
func GetRefreshTokenCollection() *mongo.Collection {
	if cachedClient == nil {
		log.Println("Database client not initialized!")
		return nil
	}
    return cachedClient.Database("propertyAppDatabase").Collection("refresh_tokens")
}
