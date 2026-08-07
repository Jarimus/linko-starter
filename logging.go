package main

import (
	"io"
	"log"
	"os"
)

func initializeLogger() *log.Logger {
	logFile := os.Getenv("LINKO_LOG_FILE")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Fatalf("failed to open log file: %v", err)
		}
		multiWriter := io.MultiWriter(file, os.Stderr)
		logger := log.New(multiWriter, "", log.LstdFlags)
		return logger
	}
	return log.New(os.Stderr, "", log.LstdFlags)

}
