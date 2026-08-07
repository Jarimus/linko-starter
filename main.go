package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"boot.dev/linko/internal/store"
)

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
	// initialize logger
	logger := initializeLogger()

	// Create store
	st, err := store.New(dataDir)
	if err != nil {
		logger.Printf("failed to create store: %v", err)
		return 1
	}

	// Create server
	s := newServer(*st, httpPort, cancel, logger)

	// Add logger to server
	s.logger = logger

	var serverErr error
	go func() {
		serverErr = s.start()
	}()

	// Wait for server to shutdown
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.logger.Println("Linko is shutting down")
	if err := s.shutdown(shutdownCtx); err != nil {
		s.logger.Printf("failed to shutdown server: %v", err)
		return 1
	}
	if serverErr != nil {
		s.logger.Printf("server error: %v", serverErr)
		return 1
	}

	return 0
}
