package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

type closeFunc func() error

func initializeLogger() (*log.Logger, closeFunc, error) {
	logFile := os.Getenv("LINKO_LOG_FILE")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %v", err)
		}
		multiWriter := bufio.NewWriterSize(io.MultiWriter(file, os.Stderr), 8192)
		logger := log.New(multiWriter, "", log.LstdFlags)
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
	return log.New(os.Stderr, "", log.LstdFlags), func() error { return nil }, nil

}
