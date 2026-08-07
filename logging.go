package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type closeFunc func() error

func initializeLogger() (*slog.Logger, closeFunc, error) {
	logFile := os.Getenv("LINKO_LOG_FILE")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %v", err)
		}
		multiWriter := bufio.NewWriterSize(io.MultiWriter(file, os.Stderr), 8192)
		logger := slog.New(slog.NewTextHandler(multiWriter, nil))
		closeLoggerFunc := func() error {
			err := multiWriter.Flush()
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
