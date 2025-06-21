package middleware

import (
	"PropertyAppBackend/config"
	"PropertyAppBackend/utils"
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

//Define a constant for the context key
const UserIDKey = "userID"


func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logrus.Warn("Authorization header missing")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logrus.Warn("Invalid Authorization header format")
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		cfg := config.GetCachedConfig()
		// Use ParseJWT to get ObjectID, but for access token, we don't need claims back here usually.
		userID, _, err := utils.ParseJWT(tokenString, cfg.JWTSecret) // ParseJWT now returns ObjectID and claims
		if err != nil {
			logrus.Warn("Invalid or expired token")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user ID (as ObjectID) to request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID) // Store ObjectID directly
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}