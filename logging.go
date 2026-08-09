package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
)

type closeFunc func() error

func initializeLogger() (*slog.Logger, closeFunc, error) {
	logFile := os.Getenv("LINKO_LOG_FILE")
	if logFile != "" {

		// Initialize debug handler
		debugHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

		// Open log file and initialize info handler
		file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %v", err)
		}
		bufWriter := bufio.NewWriterSize(file, 8192)
		infoHandler := slog.NewJSONHandler(bufWriter, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})

		// Initialize logger
		logger := slog.New(slog.NewMultiHandler(debugHandler, infoHandler))

		// Build logger closer function
		closeLoggerFunc := func() error {
			err := bufWriter.Flush()
			if err != nil {
				return fmt.Errorf("failed to flush logger buffer: %v", err)
			}
			err = file.Close()
			if err != nil {
				return fmt.Errorf("failed to close log file: %v", err)
			}
			return nil
		}
		return logger, closeLoggerFunc, nil
	}
	return slog.New(slog.NewTextHandler(os.Stderr, nil)), func() error { return nil }, nil
}
