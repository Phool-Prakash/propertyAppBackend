package database

import (
	"context"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Ctx = context.Background()

func ConnectDB(uri string) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(Ctx, 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// GetUserCollection returns the users collection
func GetUserCollection(client *mongo.Client) *mongo.Collection {
	return client.Database("propertyAppDatabase").Collection("users")
}

// GetOTPCollection returns the otps collection
func GetOTPCollection(client *mongo.Client) *mongo.Collection {
	return client.Database("propertyAppDatabase").Collection("otps")
}

func GetRefreshTokenCollection(client *mongo.Client) *mongo.Collection { // <-- NEW
	return client.Database("propertyAppDatabase").Collection("refresh_tokens")
}