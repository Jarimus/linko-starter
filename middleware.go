package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const logContextKey contextKey = "log_context"

type LogContext struct {
	Username string
	Error    error
}

type spyReadCloser struct {
	io.ReadCloser
	bytesRead int
}

func (r *spyReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.bytesRead += n
	return n, err
}

type spyResponseWriter struct {
	http.ResponseWriter
	bytesWritten int
	statusCode   int
}

func (w *spyResponseWriter) Write(p []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytesWritten += n
	return n, err
}

func (w *spyResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func redactIP(ip string) string {
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		return ip
	}
	hostSlice := net.ParseIP(host).To4()
	if hostSlice == nil {
		return ip
	}
	return fmt.Sprintf("%d.%d.%d.x", hostSlice[0], hostSlice[1], hostSlice[2])
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Insert loggingware
			startTime := time.Now()
			spyReader := &spyReadCloser{ReadCloser: r.Body}
			r.Body = spyReader
			logContext := &LogContext{}
			r = r.WithContext(context.WithValue(r.Context(), logContextKey, logContext))
			spyWriter := &spyResponseWriter{ResponseWriter: w}
			// Let request resolve
			next.ServeHTTP(spyWriter, r)
			// Log the request and response
			attrs := []any{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("client_ip", redactIP(r.RemoteAddr)),
				slog.Duration("duration", time.Since(startTime)),
				slog.Int("request_body_bytes", spyReader.bytesRead),
				slog.Int("response_status", spyWriter.statusCode),
				slog.Int("response_body_bytes", spyWriter.bytesWritten),
				slog.String("request_id", spyWriter.Header().Get("X-Request-ID")),
			}
			if logContext.Username != "" {
				attrs = append(attrs, slog.String(string(UserContextKey), logContext.Username))
			}
			if logContext.Error != nil {
				attrs = append(attrs, slog.Any("error", logContext.Error))
			}
			logger.Info("Served request", attrs...)
		})
	}
}

func requestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestIdHeader := r.Header.Get("X-Request-ID")
			if requestIdHeader == "" {
				requestIdHeader = rand.Text()
			}
			w.Header().Set("X-Request-ID", requestIdHeader)
			next.ServeHTTP(w, r)
		})
	}
}
