package config

import (
	"log"
	"os"
	"strconv"
	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	MongoDBURI             string
	JWTSecret              string // For Access Token
	RefreshTokenSecret     string // NEW: For Refresh Token
	RefreshTokenLifetimeHours int // NEW: Refresh Token lifetime
	TwilioAccountSID       string
	TwilioAuthToken        string
	TwilioPhoneNumber      string
	OTPLifetimeMinutes     int
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, assuming environment variables are set")
	}

	return &Config{
		Port:                   getEnv("PORT", ":8080"),
		MongoDBURI:             getEnv("MONGODB_URI", "mongodb://localhost:27017/propertyAppDatabase"), // Updated default DB name
		JWTSecret:              getEnv("JWT_SECRET", "supersecretjwtkey"),
		RefreshTokenSecret:     getEnv("REFRESH_TOKEN_SECRET", "another_supersecret_refresh_jwt_key"), // Default for refresh
		RefreshTokenLifetimeHours: parseIntEnv("REFRESH_TOKEN_LIFETIME_HOURS", 720), // Default 30 days
		TwilioAccountSID:       getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:        getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioPhoneNumber:      getEnv("TWILIO_PHONE_NUMBER", ""),
		OTPLifetimeMinutes:     parseIntEnv("OTP_LIFETIME_MINUTES",2),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}