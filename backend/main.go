package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
	
	// Initialize API handlers
	roomHandler := api.NewRoomHandler(roomManager)
	roomHandler.RegisterRoutes(router)
	
	// Initialize health monitoring
	healthHandler := api.NewHealthHandler(roomManager, dictionary)
	healthHandler.RegisterRoutes(router)
	
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}