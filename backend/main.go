package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"worduel-backend/internal/api"
	"worduel-backend/internal/game"
	"worduel-backend/internal/room"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := mux.NewRouter()
	
	// Initialize dictionary
	dictionary := game.NewDictionary()
	log.Printf("Dictionary loaded: %d common words, %d valid words", 
		dictionary.GetCommonWordsCount(), dictionary.GetValidWordsCount())
	
	// Initialize room manager
	roomManager := room.NewRoomManager()
	
	// Setup API middleware with CORS, rate limiting, and security
	allowedOrigins := []string{"http://localhost:3000"}
	if customOrigins := os.Getenv("ALLOWED_ORIGINS"); customOrigins != "" {
		// Allow environment variable override for production
		allowedOrigins = []string{customOrigins}
	}
	
	apiMiddleware := api.NewAPIMiddleware(allowedOrigins)
	
	// Initialize API handlers
	roomHandler := api.NewRoomHandler(roomManager)
	roomHandler.RegisterRoutes(router)
	
	// Initialize health monitoring (pass middleware for stats)
	healthHandler := api.NewHealthHandler(roomManager, dictionary, apiMiddleware)
	healthHandler.RegisterRoutes(router)
	
	// Apply middleware to all routes
	handler := apiMiddleware.ApplyMiddlewares(router)

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}