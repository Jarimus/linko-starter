package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"boot.dev/linko/internal/store"
)

var logger *log.Logger = log.New(os.Stderr, "DEBUG: ", log.LstdFlags)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	const PORT int = 8899
	httpPort := flag.Int("port", PORT, "port to listen on")
	dataDir := flag.String("data", "./data", "directory to store data")
	flag.Parse()

	status := run(ctx, cancel, *httpPort, *dataDir)
	cancel()
	os.Exit(status)
}

func run(ctx context.Context, cancel context.CancelFunc, httpPort int, dataDir string) int {
	// Create standard logger
	// stdLogger := log.New(os.Stderr, "DEBUG: ", log.LstdFlags)

	// Create store
	st, err := store.New(dataDir)
	if err != nil {
		logger.Printf("failed to create store: %v", err)
		return 1
	}

	// Create server
	s := newServer(*st, httpPort, cancel)

	// Add logger to server
	// log_file, err := os.OpenFile("linko.access.log", os.O_CREATE|os.O_WRONLY, 0666)
	// if err != nil {
	// stdLogger.Print("failed to initialize server access log file")
	// }
	// accessLogger := log.New(log_file, "INFO: ", log.LstdFlags)
	// s.logger = accessLogger
	// s.logger.Print("Server access logger initialized successfully")

	var serverErr error
	go func() {
		serverErr = s.start()
	}()

	// Wait for server to shutdown
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Println("Linko is shutting down")
	if err := s.shutdown(shutdownCtx); err != nil {
		logger.Printf("failed to shutdown server: %v", err)
		return 1
	}
	if serverErr != nil {
		logger.Printf("server error: %v", serverErr)
		return 1
	}

	return 0
}
