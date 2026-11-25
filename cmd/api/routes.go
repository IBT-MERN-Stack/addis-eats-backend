package main

import "net/http"

type middleware func(http.Handler) http.Handler

// chain applies middleware functions in reverse so the first element runs outermost [7, 23].
func chain(h http.Handler, m ...middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

// routes configures endpoints and wraps the router in the documented middleware chain [4, 22].
//
// EXECUTION ORDER (Outermost -> Innermost):
// 1. recoverPanic  : Catches panics raised anywhere in downstream wrappers/handlers [3, 4].
// 2. requestID     : Injects correlation ID into context for logging and error envelopes [4, 12].
// 3. logRequest    : Measures duration and records final HTTP status code [4, 8].
// 4. secureHeaders : Attaches X-Content-Type-Options, X-Frame-Options, etc [4, 16].
// 5. enableCORS    : Evaluates Origin whitelist and answers OPTIONS preflights [4, 14, 15].
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", app.health)
	mux.HandleFunc("GET /v1/dishes", app.listDishes)
	mux.HandleFunc("GET /v1/dishes/{id}", app.showDish)
	mux.HandleFunc("POST /v1/orders", app.createOrder)
	mux.HandleFunc("GET /v1/orders/{id}", app.showOrder)
	mux.HandleFunc("PATCH /v1/orders/{id}", app.patchOrder)
	mux.HandleFunc("/", app.notFoundRoute)

	return chain(
		mux,
		app.recoverPanic,
		app.requestID,
		app.logRequest,
		app.secureHeaders,
		app.enableCORS,
	)
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











