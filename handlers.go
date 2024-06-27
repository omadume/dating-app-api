package main

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("secret_key")
var genderOptions = []string{"female", "male", "z"}

// Custom struct to combine UserID with standard JWT claims fields & methods
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

func calculateAge(dob time.Time) int {
	now := time.Now()

	age := now.Year() - dob.Year()

	// Check if the birthday has occurred yet relative to the current date this year, and adjust age if necessary
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}

	return age
}

func filterByAge(users []User, minAge, maxAge int) []User {
	filteredUsers := []User{}
	for _, user := range users {
		age := calculateAge(user.DOB)
		if age >= minAge && age <= maxAge {
			filteredUsers = append(filteredUsers, user)
		}
	}
	return filteredUsers
}

func filterByGender(users []User, gender string) []User {
	if gender == "" {
		return users
	}
	filteredUsers := []User{}
	for _, user := range users {
		if user.Gender == gender {
			filteredUsers = append(filteredUsers, user)
		}
	}
	return filteredUsers
}

// Calculates the distance between two points using Haversine formula
func calculateDistance(targetUserLat, targetUserLon, currentUserLat, currentUserLon float64) float64 {
	const earthRadius = 6371 // Earth radius in kilometers

	// Converting latitude and longitude from degrees to radians, and calculating differences in latitude and longitude
	dLat := (currentUserLat - targetUserLat) * (math.Pi / 180)
	dLon := (currentUserLon - targetUserLon) * (math.Pi / 180)

	// Calculating Haversine formula components
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(targetUserLat*(math.Pi/180))*math.Cos(currentUserLat*(math.Pi/180))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadius * c
	return distance
}

func filterByDistance(users []User, currentUser *User, maxDistance float64) []User {
	filteredUsers := []User{}
	for _, user := range users {
		if calculateDistance(user.Latitude, user.Longitude, currentUser.Latitude, currentUser.Longitude) <= maxDistance {
			filteredUsers = append(filteredUsers, user)
		}
	}
	return filteredUsers
}

func CreateUser(responseWriter http.ResponseWriter, request *http.Request) {
	// Allowing user ages 18-90
	currentYear := time.Now().Year()
	minBirthYear := currentYear - 90
	maxBirthYear := currentYear - 18

	// Generating random user data
	user := User{
		Email:     "user" + strconv.Itoa(rand.Intn(1000)) + "@example.com",
		Password:  "password",                                 // In a real world scenario, password would be hashed/encrypted
		Name:      "username" + strconv.Itoa(rand.Intn(1000)), // Using username here for random data but would actually be full name in real case
		Gender:    genderOptions[rand.Intn(len(genderOptions))],
		DOB:       time.Date(minBirthYear+rand.Intn(maxBirthYear-minBirthYear+1), time.January, 1+rand.Intn(31), 0, 0, 0, 0, time.UTC),
		Latitude:  rand.Float64()*180 - 90,
		Longitude: rand.Float64()*360 - 180,
	}

	// Create new user
	if err := db.Create(&user).Error; err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Failed to create new user"})
		return
	}

	age := calculateAge(user.DOB)

	// Generate response
	json.NewEncoder(responseWriter).Encode(map[string]interface{}{
		"result": map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"password": user.Password,
			"name":     user.Name,
			"gender":   user.Gender,
			"age":      age,
		},
	})
}

func Login(responseWriter http.ResponseWriter, request *http.Request) {
	var userCredentials struct {
		Email    string `json:"email"` // Mapping json request keys to struct fields using tags so that it's case-insensitive matching
		Password string `json:"password"`
	}

	// Process request payload
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&userCredentials); err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Invalid request payload. Please re-check values using the docs"})
		return
	}

	// Checking for existing user using credentials
	var user User
	if err := db.Where("email = ? AND password = ?", userCredentials.Email, userCredentials.Password).First(&user).Error; err != nil {
		responseWriter.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Invalid user credentials"})
		return
	}

	// Setting 24-hr expiration for authentication token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Create authentication token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Could not generate user authentication token"})
		return
	}

	// Generate response
	json.NewEncoder(responseWriter).Encode(map[string]string{
		"token": tokenString,
	})
}

func Discover(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(USER_ID_KEY).(uint) // Retrieving userID from request context instead of decoding request payload each time

	// Find current user
	var currentUser User
	if err := db.First(&currentUser, userID).Error; err != nil {
		responseWriter.WriteHeader(http.StatusNotFound)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Could not find current user"})
		return
	}

	// Find all other users
	var users []User
	if err := db.Where("id != ?", userID).Find(&users).Error; err != nil {
		responseWriter.WriteHeader(http.StatusNotFound)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Could not find potential matches"})
		return
	}

	// Parse query parameters for filter data
	minAgeStr := request.URL.Query().Get("min_age")
	maxAgeStr := request.URL.Query().Get("max_age")
	gender := request.URL.Query().Get("gender")
	maxDistanceStr := request.URL.Query().Get("max_distance")

	// Set default values for min and max age if none or only one provided
	if minAgeStr == "" {
		minAgeStr = "18"
	}

	if maxAgeStr == "" {
		maxAgeStr = "90"
	}

	// Validating filter values and apply filtering - TODO: Could perhaps extract validation section for future improvement
	minAge, errMin := strconv.Atoi(minAgeStr)
	maxAge, errMax := strconv.Atoi(maxAgeStr)
	if errMin != nil || errMax != nil || minAge < 0 || maxAge < 0 {
		responseWriter.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Unable to process age filter value. Please enter values between 18-90"})
		return
	} else {
		users = filterByAge(users, minAge, maxAge)
	}

	if gender != "" {
		genderValid := false
		for _, option := range genderOptions {
			if option == gender {
				genderValid = true
			}
		}
		if !genderValid {
			responseWriter.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Unable to process gender filter value. Please enter either female, male, or z"})
			return
		} else {
			users = filterByGender(users, gender)
		}
	}

	if maxDistanceStr != "" {
		maxDistance, err := strconv.ParseFloat(maxDistanceStr, 64)
		if err != nil || maxDistance < 0 {
			responseWriter.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Unable to process distance filter value. Please enter a positive numeric value"})
			return
		} else {
			users = filterByDistance(users, &currentUser, maxDistance)
		}
	}

	results := []map[string]interface{}{}
	for _, user := range users {
		age := calculateAge(user.DOB)
		distance := calculateDistance(user.Latitude, user.Longitude, currentUser.Latitude, currentUser.Longitude)
		roundedDistance := int(math.Round(distance))
		results = append(results, map[string]interface{}{
			"id":             user.ID,
			"name":           user.Name,
			"gender":         user.Gender,
			"age":            age,
			"distanceFromMe": roundedDistance, // in kilometers
		})
	}

	// Set results to an empty array if no results (after filtering, etc)
	if len(results) == 0 {
		results = []map[string]interface{}{}
	}

	// Generate response - empty array returned if no potential matches
	json.NewEncoder(responseWriter).Encode(map[string]interface{}{
		"results": results,
	})
}

func Swipe(responseWriter http.ResponseWriter, request *http.Request) {
	var swipeData struct {
		TargetUserID uint   `json:"targetUserId"` // Mapping json request keys to struct fields using tags so that it's case-insensitive matching
		Preference   string `json:"preference"`
	}

	// Process request payload
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&swipeData); err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Invalid request payload. Please re-check values using the docs"})
		return
	}

	// Validating request body values - TODO: Could perhaps extract validation section for future improvement
	swipeData.Preference = strings.ToLower(swipeData.Preference)

	if swipeData.Preference != "yes" && swipeData.Preference != "no" {
		responseWriter.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Invalid preference value. Please enter either YES or NO"})
		return
	}

	userID := request.Context().Value(USER_ID_KEY).(uint) // Retrieving userID from request context instead of decoding request payload each time
	if swipeData.TargetUserID == userID {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Cannot swipe your own user"})
		return
	}

	var targetUser User
	if err := db.First(&targetUser, swipeData.TargetUserID).Error; err != nil {
		responseWriter.WriteHeader(http.StatusNotFound)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Could not find target user"})
		return
	}

	// Set up new swipe pair
	currentSwipe := SwipePair{
		UserID:       userID,
		TargetUserID: swipeData.TargetUserID,
		Preference:   swipeData.Preference,
	}

	// Create new swipe
	if err := db.Create(&currentSwipe).Error; err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Failed to create new swipe"})
		return
	}

	// Check for match between this new swipe and an existing swipe
	var existingSwipe SwipePair
	matched := false
	if currentSwipe.Preference == "yes" {
		if err := db.Where("user_id = ? AND target_user_id = ? AND preference = ?", currentSwipe.TargetUserID, userID, currentSwipe.Preference).First(&existingSwipe).Error; err == nil {
			matched = true
			existingSwipe.Match = true
			currentSwipe.Match = true
			errExistingSwipe := db.Save(&existingSwipe).Error
			errCurrentSwipe := db.Save(&currentSwipe).Error
			if errExistingSwipe != nil || errCurrentSwipe != nil {
				responseWriter.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(responseWriter).Encode(map[string]string{"error": "Failed to save match"})
				return
			}
		}
	}

	// Generate response
	response := map[string]interface{}{
		"matched": matched,
	}
	if matched {
		response["matchID"] = existingSwipe.ID
	}
	json.NewEncoder(responseWriter).Encode(map[string]interface{}{
		"results": response, // Will only contain matchID if there is a match
	})
}
