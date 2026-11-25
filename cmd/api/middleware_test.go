package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverPanic(t *testing.T) {
	app := &application{logger: log.New(io.Discard, "", 0)}
	panickingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("simulated database crash")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/dishes", nil)

	app.recoverPanic(panickingHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want 500", rr.Code)
	}
	if rr.Header().Get("Connection") != "close" {
		t.Errorf("got Connection header %q, want 'close'", rr.Header().Get("Connection"))
	}
}

func TestSecureHeaders(t *testing.T) {
	app := &application{}
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/dishes", nil)

	app.secureHeaders(nextHandler).ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("got X-Content-Type-Options %q, want 'nosniff'", got)
	}
	if got := rr.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("got X-Frame-Options %q, want 'DENY'", got)
	}
}

func TestEnableCORSPreflight(t *testing.T) {
	app := &application{
		config: config{origins: []string{"http://localhost:3000"}},
	}
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called on OPTIONS preflight")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/v1/orders", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")

	app.enableCORS(nextHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("got status %d, want 204", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("got Access-Control-Allow-Origin %q, want 'http://localhost:3000'", got)
	}
}