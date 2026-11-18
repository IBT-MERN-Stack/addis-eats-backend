package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// system
	mux.HandleFunc("GET /health", app.health)

	// menu
	mux.HandleFunc("GET /menu", app.listDishes)
	mux.HandleFunc("GET /menu/{id}", app.showDish)

	// orders
	mux.HandleFunc("POST /orders", app.createOrder)
	mux.HandleFunc("GET /orders/{id}", app.showOrder)

	
	// catch-all unmatched routes to guarantee JSON error response
	mux.HandleFunc("/", app.notFoundRoute)

	return mux
}

func (app *application) notFoundRoute(w http.ResponseWriter, r *http.Request) {
	app.notFound(w, "the requested resource was not found")
}

func (app *application) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, envelope{
		"data": map[string]string{
			"status":      "available",
			"environment": "development",
		},
	})
}