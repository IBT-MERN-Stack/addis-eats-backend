package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/menu"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/order"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/delivery"
)

type createOrderRequest struct {
	Area    string   `json:"area"`
	Phone   string   `json:"phone"`
	DishIDs []string `json:"dish_ids"`
}

func (app *application) createOrder(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)
	var in createOrderRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
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

	w.Header().Set("Location", "/orders/"+o.ID)
	writeJSON(w, http.StatusCreated, envelope{"data": o})
}

// showOrder implements the parse-delegate-respond pattern.
func (app *application) showOrder(w http.ResponseWriter, r *http.Request) {
	// 1. Parse
	id, err := readID(r)
	if err != nil {
		app.notFound(w, "order not found")
		return
	}

	// 2. Delegate
	o, err := app.orders.ByID(id)

	// 3. Map and Respond
	switch {
	case errors.Is(err, order.ErrNotFound):
		app.notFound(w, "order not found")
	case err != nil:
		app.serverError(w, err)
	default:
		writeJSON(w, http.StatusOK, envelope{"data": o})
	}
}