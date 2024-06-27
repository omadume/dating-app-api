package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	InitDatabase()

	router := mux.NewRouter()

	// Setting up valid API routes
	router.Handle("/user/create", ValidateQueryParams(http.HandlerFunc(CreateUser))).Methods("POST") // Have defined this as POST as it would be in a real world scenario, although in this case creates random user.
	router.Handle("/login", ValidateQueryParams(http.HandlerFunc(Login))).Methods("POST")
	router.Handle("/discover", ValidateQueryParams(http.HandlerFunc(Authenticate(Discover)))).Methods("GET")
	router.Handle("/swipe", ValidateQueryParams(http.HandlerFunc(Authenticate(Swipe)))).Methods("POST")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
