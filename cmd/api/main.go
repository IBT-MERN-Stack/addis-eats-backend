package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/menu"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/order"
)

type config struct {
	origins []string
}

type application struct {
	config config
	logger *log.Logger
	dishes menu.DishReader
	orders order.Store
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Load allowed CORS origins from environment variable (default to dev frontend) [16]
	rawOrigins := os.Getenv("CORS_ORIGINS")
	if rawOrigins == "" {
		rawOrigins = "http://localhost:3000"
	}
	origins := strings.Split(rawOrigins, ",")

	app := &application{
		config: config{origins: origins},
		logger: logger,
		dishes: menu.NewMemoryStore(),
		orders: order.NewMemoryStore(),
	}

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      app.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	logger.Printf("starting server on %s", srv.Addr)
	logger.Fatal(srv.ListenAndServe())
}