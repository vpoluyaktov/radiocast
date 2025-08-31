package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"radiocast/internal/config"
	"radiocast/internal/server"
	"radiocast/internal/testing"
)

func main() {
	ctx := context.Background()
	
	// Parse command line flags
	deploymentFlag := flag.String("deployment", "local", "Deployment mode: local or gcs")
	testChartsFlag := flag.Bool("test-charts", false, "Generate test charts and exit")
	flag.Parse()
	
	// Validate deployment mode
	var deploymentMode server.DeploymentMode
	switch *deploymentFlag {
	case "local":
		deploymentMode = server.DeploymentLocal
	case "gcs":
		deploymentMode = server.DeploymentGCS
	default:
		log.Fatalf("Invalid deployment mode: %s. Use 'local' or 'gcs'", *deploymentFlag)
	}
	
	// Load configuration
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Handle test charts mode (skip config validation for chart testing)
	if *testChartsFlag {
		testing.RunTestCharts()
		return
	}
	
	log.Printf("Starting Radio Propagation Service on port %s", cfg.Port)
	log.Printf("Deployment mode: %s", deploymentMode)
	
	// Create server
	srv, err := server.NewServer(cfg, deploymentMode)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()
	
	// Set up HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.HandleRoot)
	mux.HandleFunc("/health", srv.HandleHealth)
	mux.HandleFunc("/generate", srv.HandleGenerate)
	mux.HandleFunc("/reports", srv.HandleListReports)
	mux.HandleFunc("/files/", srv.HandleFileProxy)
	
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
		log.Printf("Server listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	log.Println("Shutting down server...")
	
	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	
	log.Println("Server stopped")
}
