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

// Wire request struct keeps client from providing prices or totals
type createOrderRequest struct {
	Area    string   `json:"area"`
	Phone   string   `json:"phone"`
	DishIDs []string `json:"dish_ids"`
}

func (app *application) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, envelope{
		"data": map[string]string{
			"status":      "available",
			"environment": "development",
		},
	})
}

func (app *application) listDishes(w http.ResponseWriter, r *http.Request) {
	dishes, err := app.dishes.All()
	if err != nil {
		app.serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": dishes})
}


func (app *application) showDish(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	dish, err := app.dishes.ByID(id)
	switch {
	case errors.Is(err, menu.ErrNotFound):
		app.notFound(w, "dish not found")
	case err != nil:
		app.serverError(w, err)
	default:
		writeJSON(w, http.StatusOK, envelope{"data": dish})
	}
}

func (app *application) createOrder(w http.ResponseWriter, r *http.Request) {
   
	// Limit body payload to 1MB
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

	// Lookup dishes from store to guarantee server-controlled pricing
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

	// Set Location header prior to writing status code
	w.Header().Set("Location", "/orders/"+o.ID)
	writeJSON(w, http.StatusCreated, envelope{"data": o})
}