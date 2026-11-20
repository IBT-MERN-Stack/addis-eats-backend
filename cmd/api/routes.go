package main

import "net/http"

// routes defines the application's central routing table.
// TEACHING NOTE (REST Naming):
// 1. Plural Nouns: Endpoints use `/v1/dishes` instead of `/v1/menu` or `/v1/createOrder` [13, 14].
// 2. Explicit Version Prefix: `/v1` guarantees backward compatibility as APIs evolve [15, 16].
// 3. Catch-All Route: Catch-all `"/"` guarantees unmapped URLs answer with JSON error envelopes.
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// System Endpoints
	mux.HandleFunc("GET /v1/health", app.health)

	// Dishes Collection & Resource Endpoints
	mux.HandleFunc("GET /v1/dishes", app.listDishes)
	mux.HandleFunc("GET /v1/dishes/{id}", app.showDish)

	// Orders Collection & Resource Endpoints
	mux.HandleFunc("POST /v1/orders", app.createOrder)
	mux.HandleFunc("GET /v1/orders/{id}", app.showOrder)
	mux.HandleFunc("PATCH /v1/orders/{id}", app.patchOrder)

	// Catch-all Unmatched Route
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
			"version":     "v1",
		},
	})
}