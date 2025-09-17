package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"radiocast/internal/config"
	"radiocast/internal/logger"
	"radiocast/internal/server"
	"radiocast/internal/storage"
)

// initializeLogger configures the global logger based on configuration and deployment mode
func initializeLogger(cfg *config.Config, deploymentMode storage.DeploymentMode) {
	// Parse log level
	var level logger.LogLevel
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = logger.DEBUG
	case "info":
		level = logger.INFO
	case "warn", "warning":
		level = logger.WARN
	case "error":
		level = logger.ERROR
	case "fatal":
		level = logger.FATAL
	default:
		level = logger.INFO
	}
	
	// Determine log format based on deployment mode and configuration
	var format logger.LogFormat
	
	switch strings.ToLower(cfg.LogFormat) {
	case "json":
		format = logger.JSONFormat
	case "text":
		format = logger.TextFormat
	case "auto":
		// Auto-detect based on deployment mode
		switch deploymentMode {
		case storage.DeploymentLocal:
			format = logger.TextFormat  // Human-readable for local development
		case storage.DeploymentGCS:
			format = logger.JSONFormat  // Structured for production/GCS
		default:
			format = logger.JSONFormat
		}
	default:
		// Default to auto-detection if unknown format
		switch deploymentMode {
		case storage.DeploymentLocal:
			format = logger.TextFormat
		case storage.DeploymentGCS:
			format = logger.JSONFormat
		default:
			format = logger.JSONFormat
		}
	}
	
	// Create and set global logger
	loggerConfig := logger.Config{
		Level:     level,
		Format:    format,
		Output:    os.Stdout,
		Component: "radiocast-service",
	}
	
	logger.SetGlobalLogger(logger.New(loggerConfig))
	
	// Log the selected configuration
	formatName := "JSON"
	if format == logger.TextFormat {
		formatName = "Text"
	}
	logger.Infof("Logger initialized - Level: %s, Format: %s (based on deployment mode: %s)", 
		strings.ToUpper(cfg.LogLevel), formatName, deploymentMode)
}

func main() {
	ctx := context.Background()
	
	// Parse command line flags
	deploymentFlag := flag.String("deployment", "local", "Deployment mode: local or gcs")
	flag.Parse()
	
	// Validate deployment mode
	var deploymentMode storage.DeploymentMode
	switch *deploymentFlag {
	case "local":
		deploymentMode = storage.DeploymentLocal
	case "gcs":
		deploymentMode = storage.DeploymentGCS
	default:
		log.Fatalf("Invalid deployment mode: %s. Use 'local' or 'gcs'", *deploymentFlag)
	}
	
	// Load configuration
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Initialize structured logger
	initializeLogger(cfg, deploymentMode)
	
	logger.Infof("Starting Radio Propagation Service on port %s", cfg.Port)
	logger.Infof("Deployment mode: %s", deploymentMode)
	
	// Create server
	srv, err := server.NewServer(cfg, deploymentMode)
	if err != nil {
		logger.Fatal("Failed to create server", err)
	}
	defer srv.Close()
	
	// Set up HTTP routes using server's routing configuration
	mux := srv.SetupRoutes()
	
	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second, // Longer timeout for report generation
		IdleTimeout:  60 * time.Second,
	}
	
	// Start server in goroutine
	go func() {
		logger.Infof("Server listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal("HTTP server error", err)
		}
	}()
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	logger.Info("Shutting down server...")
	
	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", err)
	}
	
	logger.Info("Server stopped")
}
