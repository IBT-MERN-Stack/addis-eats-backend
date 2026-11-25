package main

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"
)

// contextKey uses an unexported custom type to prevent collision with third-party context keys [10, 11].
type contextKey string

const requestIDKey contextKey = "requestID"

// requestIDFrom extracts the request correlation ID from the request context [11].
func requestIDFrom(r *http.Request) string {
	id, ok := r.Context().Value(requestIDKey).(string)
	if !ok {
		return "unknown"
	}
	return id
}

// statusWriter wraps http.ResponseWriter via struct embedding to capture the HTTP status code [8, 9].
type statusWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader shadows the embedded method to record the written HTTP status code [8, 9].
func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// 1. recoverPanic MUST be outermost to catch panics anywhere in the chain [3, 4].
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Signal the HTTP server to close the connection after responding [19, 20]
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("panic: %v", rec))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// 2. requestID assigns or sanitizes X-Request-Id and stores it in the Context [12, 21].
func (app *application) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = newID() // Generates crypto-random ID [21]
		} else {
			// Cap length and sanitize untrusted incoming request IDs [12]
			if len(id) > 64 {
				id = id[:64]
			}
		}
		w.Header().Set("X-Request-Id", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx)) // Pass new request containing Context [11, 21]
	})
}

// 3. logRequest logs request method, path, HTTP status, duration, and correlation ID [8].
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r) // Pass wrapped statusWriter down the chain [8, 9]
		app.logger.Printf(
			"%s %s %d %s %s",
			r.Method,
			r.URL.Path,
			sw.status,
			time.Since(start).Round(time.Millisecond),
			requestIDFrom(r),
		)
	})
}

// 4. secureHeaders sets baseline HTTP security headers before writing the response [7, 16].
func (app *application) secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// 5. enableCORS enforces origin whitelisting and short-circuits OPTIONS preflights [14, 15].
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin") // Informs caches that responses depend on Origin [14, 15]
		origin := r.Header.Get("Origin")
		if origin != "" && app.config.origins != nil && slices.Contains(app.config.origins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			// Handle CORS preflight OPTIONS request [13, 14]
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id")
				w.Header().Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent) // Return 204 No Content [13, 14]
				return                        // Short-circuit without calling next [10, 15]
			}
		}
		next.ServeHTTP(w, r)
	})
}




