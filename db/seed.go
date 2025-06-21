package database

import (
    "context"
    "log"
    "PropertyAppBackend/models"
    "PropertyAppBackend/utils"

    "go.mongodb.org/mongo-driver/bson"
)

// Ensure Default Admin Exists
func SeedAdminUser() {
    collection := GetUserCollection()

    // **Check if Admin already exists**
    adminCount, _ := collection.CountDocuments(context.Background(), bson.M{"role": models.Admin})
    if adminCount > 0 {
        log.Println("Admin already exists, skipping seed.")
        return
    }

    // **Create Default Admin**
	adminUsername := "admin"
    hashedPassword, _ := utils.HashPassword("default_admin_password") 
    admin := models.User{
        Username: &adminUsername,
        Password: &hashedPassword,
        Role:     models.Admin,
    }

    _, err := collection.InsertOne(context.Background(), admin)
    if err != nil {
        log.Println("Error seeding Admin user:", err)
    }
}
