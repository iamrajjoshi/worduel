package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"worduel-backend/internal/api"
	"worduel-backend/internal/config"
	"worduel-backend/internal/game"
	"worduel-backend/internal/logging"
	"worduel-backend/internal/room"
	"worduel-backend/internal/ws"

	"github.com/gorilla/mux"
)

type Application struct {
	config      *config.Config
	server      *http.Server
	hub         *ws.Hub
	roomManager *room.RoomManager
	dictionary  *game.Dictionary
	gameLogic   *game.GameLogic
	logger      *logging.Logger
}

func main() {
	app := &Application{}

	if err := app.Initialize(); err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		app.logger.Error("Application failed", "error", err.Error())
		os.Exit(1)
	}
}

func (app *Application) Initialize() error {
	// Load and validate configuration first (before we can set up logging)
	if err := app.loadConfiguration(); err != nil {
		return fmt.Errorf("configuration loading failed: %w", err)
	}

	// Initialize logging and Sentry
	if err := app.initializeLogging(); err != nil {
		return fmt.Errorf("logging initialization failed: %w", err)
	}

	app.logger.Info("Initializing application...")

	// Initialize core components
	if err := app.initializeComponents(); err != nil {
		app.logger.Error("Component initialization failed", "error", err.Error())
		return fmt.Errorf("component initialization failed: %w", err)
	}

	// Set up HTTP server and routes
	if err := app.setupServer(); err != nil {
		app.logger.Error("Server setup failed", "error", err.Error())
		return fmt.Errorf("server setup failed: %w", err)
	}

	app.logger.Info("Application initialized successfully")
	return nil
}

func (app *Application) loadConfiguration() error {
	fmt.Println("Loading configuration...")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	app.config = cfg
	fmt.Printf("Configuration loaded - Server: %s:%s\n", cfg.Server.Host, cfg.Server.Port)

	if cfg.Dev.DebugMode {
		fmt.Println("Running in debug mode")
	}

	return nil
}

func (app *Application) initializeLogging() error {
	fmt.Println("Initializing logging...")

	// Initialize Sentry if DSN is provided
	dsn := app.config.Sentry.DSN
	if dsn == "" {
		dsn = "https://8c7abeba76bbc23136f9f0284e7e8e02@o4509294838415360.ingest.us.sentry.io/4509888598769664"
	}

	sentryConfig := logging.SentryConfig{
		DSN:              dsn,
		Environment:      app.config.Sentry.Environment,
		Release:          app.config.Sentry.Release,
		TracesSampleRate: app.config.Sentry.TracesSampleRate,
		Debug:            app.config.Sentry.Debug,
	}

	if err := logging.InitSentry(sentryConfig); err != nil {
		return fmt.Errorf("failed to initialize Sentry: %w", err)
	}
	fmt.Println("Sentry initialized")

	// Initialize structured logger
	loggerConfig := logging.LogConfig{
		Level:       app.config.Logging.Level,
		Environment: app.config.Logging.Environment,
		Service:     app.config.Logging.Service,
		AddSource:   app.config.Logging.AddSource,
	}

	logger, err := logging.NewLogger(loggerConfig)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	app.logger = logger
	logging.SetGlobalLogger(logger)
	fmt.Println("Structured logging initialized")

	return nil
}

func (app *Application) initializeComponents() error {
	app.logger.Info("Initializing core components...")

	// Initialize dictionary service
	app.dictionary = game.NewDictionary()
	app.logger.Info("Dictionary loaded",
		"common_words", app.dictionary.GetCommonWordsCount(),
		"valid_words", app.dictionary.GetValidWordsCount())

	// Initialize room manager with configuration
	app.roomManager = room.NewRoomManager()
	app.roomManager.SetMaxConcurrentRooms(app.config.Room.MaxConcurrentRooms)
	app.logger.Info("Room manager initialized", "max_rooms", app.config.Room.MaxConcurrentRooms)

	// Initialize game logic
	app.gameLogic = game.NewGameLogic(app.dictionary)
	app.logger.Info("Game logic initialized")

	// Initialize WebSocket hub
	app.hub = ws.NewHub(app.roomManager, app.gameLogic)
	app.logger.Info("WebSocket hub initialized")

	return nil
}

func (app *Application) setupServer() error {
	app.logger.Info("Setting up HTTP server and routes...")

	router := mux.NewRouter()

	// Setup API middleware with configuration
	apiMiddleware := api.NewAPIMiddleware(app.config.CORS.AllowedOrigins)

	// Add Sentry HTTP middleware if configured
	if app.config.Sentry.DSN != "" {
		router.Use(logging.SentryHTTPMiddleware())
		app.logger.Info("Sentry HTTP middleware configured")
	}

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

	app.logger.Info("HTTP server configured", "address", app.server.Addr)
	return nil
}

func (app *Application) Run() error {
	// Start background services
	app.startBackgroundServices()

	// Start HTTP server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		app.logger.Info("Server starting", "address", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Wait for interrupt signal or server error
	return app.waitForShutdownSignal(serverErrChan)
}

func (app *Application) startBackgroundServices() {
	app.logger.Info("Starting background services...")

	// Start WebSocket hub
	go app.hub.Run()
	app.logger.Info("WebSocket hub started")

	// Start room cleanup service
	go app.startRoomCleanup()
	app.logger.Info("Room cleanup service started")
}

func (app *Application) startRoomCleanup() {
	ticker := time.NewTicker(app.config.Room.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		cleanedCount := app.roomManager.CleanupExpiredRooms(app.config.Room.RoomInactiveTimeout)
		if cleanedCount > 0 {
			app.logger.Info("Cleaned up expired rooms", "count", cleanedCount)
		}
	}
}

func (app *Application) waitForShutdownSignal(serverErrChan chan error) error {
	// Create channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		app.logger.Error("Server error", "error", err.Error())
		return err
	case sig := <-quit:
		app.logger.Info("Received shutdown signal", "signal", sig)
		return app.gracefulShutdown()
	}
}

func (app *Application) gracefulShutdown() error {
	app.logger.Info("Starting graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.config.Server.ShutdownTimeout)
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Shutdown HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.logger.Info("Shutting down HTTP server...")
		if err := app.server.Shutdown(shutdownCtx); err != nil {
			errChan <- fmt.Errorf("server shutdown failed: %w", err)
			return
		}
		app.logger.Info("HTTP server stopped")
	}()

	// Shutdown WebSocket hub
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.logger.Info("Shutting down WebSocket hub...")
		app.hub.Shutdown()
		app.logger.Info("WebSocket hub stopped")
	}()

	// Cleanup room manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.logger.Info("Cleaning up room manager...")
		app.roomManager.Shutdown()
		app.logger.Info("Room manager stopped")
	}()

	// Wait for all shutdown operations or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		app.logger.Info("Graceful shutdown completed successfully")
		// Flush Sentry before exit
		logging.FlushSentry(2 * time.Second)
		return nil
	case err := <-errChan:
		app.logger.Error("Shutdown error", "error", err.Error())
		logging.FlushSentry(2 * time.Second)
		return err
	case <-shutdownCtx.Done():
		app.logger.Warn("Shutdown timeout exceeded, forcing exit")
		logging.FlushSentry(2 * time.Second)
		return fmt.Errorf("shutdown timeout exceeded")
	}
}
