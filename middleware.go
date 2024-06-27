package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// Custom type for context keys
// - using built-in string type directly could lead to potential collisions with context keys defined by other packages / modules within app in future
type contextKey string

const USER_ID_KEY contextKey = "userID"

// Middleware function to validate query parameters for the defined routes
func ValidateQueryParams(handlerFunction http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		// Accepted query parameters for each route - currently only /discover allows query params
		acceptedParams := map[string]map[string]bool{
			"/user/create": {},
			"/login":       {},
			"/discover":    {"min_age": true, "max_age": true, "gender": true, "max_distance": true},
			"/swipe":       {},
		}

		// Parse request URL to get the path for the specific route
		path := request.URL.Path

		// Check if the path is in the acceptedParams map, and check its associate query params (if any) are valid
		if params, ok := acceptedParams[path]; ok {
			// Parse query parameters
			query := request.URL.Query()

			// Check for invalid query parameters
			for param := range query {
				if !params[param] {
					responseWriter.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(responseWriter).Encode(map[string]string{"error": "URL contains invalid query parameter"})
					return
				}
			}
		}

		handlerFunction.ServeHTTP(responseWriter, request)
	})
}

// Middleware function to authenticate particular routes
func Authenticate(handlerFunction http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		// Retrieve authorization token from request header
		authHeader := request.Header.Get("Authorization")
		if authHeader == "" {
			responseWriter.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Request missing authorization header"})
			return
		}

		tokenStringParts := strings.Split(authHeader, " ")
		if len(tokenStringParts) < 2 {
			responseWriter.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(responseWriter).Encode(map[string]string{"error": "No authentication token value provided"})
			return
		}

		tokenString := tokenStringParts[1]
		claims := &Claims{}

		// Parse authentication token to authenticate
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			responseWriter.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Could not authenticate token"})
			return
		}

		// Storing userID in request context for use in HTTP handler functions that are passed through Authenticate()
		ctx := context.WithValue(request.Context(), USER_ID_KEY, claims.UserID)
		handlerFunction.ServeHTTP(responseWriter, request.WithContext(ctx))
	})
}
