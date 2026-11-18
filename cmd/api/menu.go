package main

import (
	"errors"
	"net/http"

	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/menu"
)

func (app *application) listDishes(w http.ResponseWriter, r *http.Request) {

	
	qs := r.URL.Query()

	fastingVal, fastingOk := readBool(qs, "fasting")
	if qs.Has("fasting") && !fastingOk {
		app.badRequest(w, "fasting must be true or false")
		return
	}

	sortVal, sortOk := readSort(qs, menu.SortFields, "name")
	if !sortOk {
		app.badRequest(w, "unknown sort: "+sortVal)
		return
	}

	q := menu.Query{Sort: sortVal}
	if qs.Has("fasting") {
		q.Fasting = &fastingVal
	}

	dishes, err := app.dishes.Filter(q)
	if err != nil {
		app.serverError(w, err)
		return
	}

	filtersMeta := map[string]any{}
	if q.Fasting != nil {
		filtersMeta["fasting"] = *q.Fasting
	}

	writeJSON(w, http.StatusOK, envelope{
		"data": dishes,
		"meta": map[string]any{
			"count":   len(dishes),
			"sort":    sortVal,
			"filters": filtersMeta,
		},
	})
}

func (app *application) showDish(w http.ResponseWriter, r *http.Request) {
	id, err := readID(r)
	if err != nil {
		app.notFound(w, "dish not found")
		return
	}

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