package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"worduel-backend/internal/api"
	"worduel-backend/internal/config"
	"worduel-backend/internal/game"
	"worduel-backend/internal/room"
	"worduel-backend/internal/ws"
)

type Application struct {
	config      *config.Config
	server      *http.Server
	hub         *ws.Hub
	roomManager *room.RoomManager
	dictionary  *game.Dictionary
	gameLogic   *game.GameLogic
	logger      *log.Logger
}

func main() {
	app := &Application{
		logger: log.New(os.Stdout, "[WORDUEL] ", log.LstdFlags|log.Lshortfile),
	}

	if err := app.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func (app *Application) Initialize() error {
	app.logger.Println("Initializing application...")

	// Load and validate configuration
	if err := app.loadConfiguration(); err != nil {
		return fmt.Errorf("configuration loading failed: %w", err)
	}

	// Initialize core components
	if err := app.initializeComponents(); err != nil {
		return fmt.Errorf("component initialization failed: %w", err)
	}

	// Set up HTTP server and routes
	if err := app.setupServer(); err != nil {
		return fmt.Errorf("server setup failed: %w", err)
	}

	app.logger.Println("Application initialized successfully")
	return nil
}

func (app *Application) loadConfiguration() error {
	app.logger.Println("Loading configuration...")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	app.config = cfg
	app.logger.Printf("Configuration loaded - Server: %s:%s", cfg.Server.Host, cfg.Server.Port)

	if cfg.Dev.DebugMode {
		app.logger.Println("Running in debug mode")
	}

	return nil
}

func (app *Application) initializeComponents() error {
	app.logger.Println("Initializing core components...")

	// Initialize dictionary service
	app.dictionary = game.NewDictionary()
	app.logger.Printf("Dictionary loaded: %d common words, %d valid words",
		app.dictionary.GetCommonWordsCount(), app.dictionary.GetValidWordsCount())

	// Initialize room manager with configuration
	app.roomManager = room.NewRoomManager()
	app.roomManager.SetMaxConcurrentRooms(app.config.Room.MaxConcurrentRooms)
	app.logger.Printf("Room manager initialized (max rooms: %d)", app.config.Room.MaxConcurrentRooms)

	// Initialize game logic
	app.gameLogic = game.NewGameLogic(app.dictionary)
	app.logger.Println("Game logic initialized")

	// Initialize WebSocket hub
	app.hub = ws.NewHub(app.roomManager, app.gameLogic)
	app.logger.Println("WebSocket hub initialized")

	return nil
}

func (app *Application) setupServer() error {
	app.logger.Println("Setting up HTTP server and routes...")

	router := mux.NewRouter()

	// Setup API middleware with configuration
	apiMiddleware := api.NewAPIMiddleware(app.config.CORS.AllowedOrigins)

	// Initialize API handlers
	roomHandler := api.NewRoomHandler(app.roomManager)
	roomHandler.RegisterRoutes(router)

	healthHandler := api.NewHealthHandler(app.roomManager, app.dictionary, apiMiddleware)
	healthHandler.RegisterRoutes(router)

	// Setup WebSocket endpoint
	wsHandler := ws.NewHandler(app.hub, app.roomManager, app.dictionary)
	router.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Apply middleware to all routes
	handler := apiMiddleware.ApplyMiddlewares(router)

	// Create HTTP server with configuration
	app.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", app.config.Server.Host, app.config.Server.Port),
		Handler:      handler,
		ReadTimeout:  app.config.Server.ReadTimeout,
		WriteTimeout: app.config.Server.WriteTimeout,
		IdleTimeout:  app.config.Server.IdleTimeout,
	}

	app.logger.Printf("HTTP server configured on %s", app.server.Addr)
	return nil
}

func (app *Application) Run() error {
	// Start background services
	app.startBackgroundServices()

	// Start HTTP server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		app.logger.Printf("Server starting on %s", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Wait for interrupt signal or server error
	return app.waitForShutdownSignal(serverErrChan)
}

func (app *Application) startBackgroundServices() {
	app.logger.Println("Starting background services...")

	// Start WebSocket hub
	go app.hub.Run()
	app.logger.Println("WebSocket hub started")

	// Start room cleanup service
	go app.startRoomCleanup()
	app.logger.Println("Room cleanup service started")
}

func (app *Application) startRoomCleanup() {
	ticker := time.NewTicker(app.config.Room.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		cleanedCount := app.roomManager.CleanupExpiredRooms(app.config.Room.RoomInactiveTimeout)
		if cleanedCount > 0 && app.config.Dev.VerboseLog {
			app.logger.Printf("Cleaned up %d expired rooms", cleanedCount)
		}
	}
}

func (app *Application) waitForShutdownSignal(serverErrChan chan error) error {
	// Create channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		app.logger.Printf("Server error: %v", err)
		return err
	case sig := <-quit:
		app.logger.Printf("Received shutdown signal: %v", sig)
		return app.gracefulShutdown()
	}
}

func (app *Application) gracefulShutdown() error {
	app.logger.Println("Starting graceful shutdown...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), app.config.Server.ShutdownTimeout)
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Shutdown HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.logger.Println("Shutting down HTTP server...")
		if err := app.server.Shutdown(ctx); err != nil {
			errChan <- fmt.Errorf("server shutdown failed: %w", err)
			return
		}
		app.logger.Println("HTTP server stopped")
	}()

	// Shutdown WebSocket hub
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.logger.Println("Shutting down WebSocket hub...")
		app.hub.Shutdown()
		app.logger.Println("WebSocket hub stopped")
	}()

	// Cleanup room manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.logger.Println("Cleaning up room manager...")
		app.roomManager.Shutdown()
		app.logger.Println("Room manager stopped")
	}()

	// Wait for all shutdown operations or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		app.logger.Println("Graceful shutdown completed successfully")
		return nil
	case err := <-errChan:
		app.logger.Printf("Shutdown error: %v", err)
		return err
	case <-ctx.Done():
		app.logger.Println("Shutdown timeout exceeded, forcing exit")
		return fmt.Errorf("shutdown timeout exceeded")
	}
}