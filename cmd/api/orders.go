package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/delivery"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/menu"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/order"
)

type createOrderRequest struct {
	Area    string   `json:"area"`
	Phone   string   `json:"phone"`
	DishIDs []string `json:"dish_ids"`
}

// TEACHING NOTE (PATCH & Pointer Fields):
// `patchOrderRequest` uses a pointer `*order.Status` rather than a raw `order.Status`.
// In Go, unmarshaling JSON into a raw struct assigns omitted fields their zero-values.
// Using pointer fields allows handlers to distinguish between:
// 1. nil => Field was omitted in the JSON request body (do not update).
// 2. &val => Field was explicitly provided (apply update) [6, 7].
type patchOrderRequest struct {
	Status *order.Status `json:"status"`
}

func (app *application) createOrder(w http.ResponseWriter, r *http.Request) {
	// Restrict payload body size to 1MB to prevent memory exhaustion attacks
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	var in createOrderRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // Reject unexpected fields to enforce API contract rigidity
	if err := dec.Decode(&in); err != nil {
		app.badRequest(w, "malformed JSON body or unknown field")
		return
	}

	if len(in.DishIDs) == 0 {
		app.badRequest(w, "dish_ids must not be empty")
		return
	}

	items := make([]menu.Dish, 0, len(in.DishIDs))
	for _, id := range in.DishIDs {
		d, err := app.dishes.ByID(id)
		if err != nil {
			app.badRequest(w, fmt.Sprintf("unknown dish %s", id))
			return
		}
		items = append(items, d)
	}

	o := order.New(in.Area, in.Phone, items)
	_, err := o.ComputeTotal()
	if err != nil {
		if errors.Is(err, delivery.ErrUnservedArea) {
			app.badRequest(w, fmt.Sprintf("we do not deliver to %s", in.Area))
			return
		}
		app.badRequest(w, err.Error())
		return
	}

	if err := app.orders.Create(&o); err != nil {
		app.serverError(w, err)
		return
	}

	// Set Location header for newly created resources (201 Created) [7]
	w.Header().Set("Location", "/v1/orders/"+o.ID)
	writeJSON(w, http.StatusCreated, envelope{"data": o})
}

func (app *application) showOrder(w http.ResponseWriter, r *http.Request) {
	id, err := readID(r)
	if err != nil {
		app.notFound(w, "order not found")
		return
	}

	o, err := app.orders.ByID(id)
	switch {
	case errors.Is(err, order.ErrNotFound):
		app.notFound(w, "order not found")
	case err != nil:
		app.serverError(w, err)
	default:
		// TEACHING NOTE: User orders are private and volatile.
		// Always attach `Cache-Control: no-store` to prevent intermediary proxies from caching sensitive customer data [4, 5].
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, envelope{"data": o})
	}
}

// patchOrder handles partial updates for orders following the parse-delegate-respond cycle [6].
func (app *application) patchOrder(w http.ResponseWriter, r *http.Request) {
	// 1. Parse Input & Validate Body
	id, err := readID(r)
	if err != nil {
		app.notFound(w, "order not found")
		return
	}

	var in patchOrderRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		app.badRequest(w, "malformed JSON payload")
		return
	}

	if in.Status == nil {
		app.badRequest(w, "status field is required for patch")
		return
	}

	// 2. Delegate Update to Store Layer
	updatedOrder, err := app.orders.UpdateStatus(id, *in.Status)

	// 3. Map Domain Errors to HTTP Statuses
	switch {
	case errors.Is(err, order.ErrNotFound):
		app.notFound(w, "order not found")
	case errors.Is(err, order.ErrIllegalTransition):
		// TEACHING NOTE (409 Conflict): Return 409 Conflict when a request is well-formed
		// but contradicts the current state of the resource [7, 8].
		app.conflict(w, err.Error())
	case err != nil:
		app.serverError(w, err)
	default:
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusNoContent, envelope{"data": updatedOrder})
	}
}