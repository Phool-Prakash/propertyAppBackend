// This file manage the application configuration variables and load the environment variables
package config

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

//this is for batter logging
var once sync.Once
var cachedCfg *Config

//Config Structure (this is encapsulate the important settings)
type Config struct {
	Port                   string
	MongoDBURI             string
	JWTSecret              string 
	RefreshTokenSecret     string 
	RefreshTokenLifetimeHours int 
	TwilioAccountSID       string
	TwilioAuthToken        string
	TwilioPhoneNumber      string
	OTPLifetimeMinutes     int
}


//LoadConfig func
func LoadConfig() *Config {
	once.Do(func ()  {
		err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, assuming environment variables are set")
		logrus.Warn("Error loading .env file, assuming environment variables are set")
	}

	cachedCfg =  &Config{
		Port:                   getEnv("PORT", ":8080"),
		MongoDBURI:             getEnv("MONGODB_URI", "mongodb://localhost:27017/propertyAppDatabase"),
		JWTSecret:              getSecureEnv("JWT_SECRET"),
		RefreshTokenSecret:     getSecureEnv("REFRESH_TOKEN_SECRET"), 
		RefreshTokenLifetimeHours: parseIntEnv("REFRESH_TOKEN_LIFETIME_HOURS", 720), // 
		TwilioAccountSID:       getSecureEnv("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:        getSecureEnv("TWILIO_AUTH_TOKEN"),
		TwilioPhoneNumber:      getSecureEnv("TWILIO_PHONE_NUMBER"),
		OTPLifetimeMinutes:     parseIntEnv("OTP_LIFETIME_MINUTES",2),
	}
	})
	return cachedCfg
}

//Retrive cached Config instance
func GetCachedConfig() *Config {
	return cachedCfg
}

//Secure environment variable retrieval
func getSecureEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	logrus.Error("Missing secure env variable:" +key)
	return ""
}

//getEnv func to fetch the env variables and if env variables value is missing then it return defaultValue (We say that it's general environment variable retrieval)
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	logrus.Warn("Using default value for env variable: " + key)
	return defaultValue
}


//The working of this func is same as the working of getEnv func but if env variable is not present in the integer format then it return defaultValue
// Parse integer environment variables with error handling
func parseIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
		logrus.Error("Invalid integer value for env variable:" + key)
	}
	return defaultValue
}

