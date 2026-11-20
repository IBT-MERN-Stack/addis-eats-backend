package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/menu"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/order"
)

type application struct {
	logger *log.Logger
	dishes menu.DishReader
	orders order.Store
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Inject dependencies into application struct
	app := &application{
		logger: logger,
		dishes: menu.NewMemoryStore(),
		orders: order.NewMemoryStore(),
	}

	// Configure server with explicit connection timeouts to prevent slowloris attacks
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