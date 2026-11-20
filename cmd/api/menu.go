package main

import (
	"errors"
	"math"
	"net/http"

	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/menu"
)

func (app *application) listDishes(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()

	// 1. Validate & Clamp Pagination Parameters [9, 10]
	page, err := readInt(qs, "page", 1)
	if err != nil || page < 1 {
		app.badRequest(w, "page must be a positive integer")
		return
	}

	limit, err := readInt(qs, "limit", 20)
	if err != nil || limit < 1 {
		app.badRequest(w, "limit must be a positive integer")
		return
	}

	// TEACHING NOTE: Always clamp collection limits using `min(limit, 100)`.
	// Unbounded limit parameters allow malicious requests to exhaust server memory [10, 17].
	limit = min(limit, 100)

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

	// 2. HTTP Caching & Conditional Verification (ETags) [4]
	etag := app.dishes.ETag(q)
	w.Header().Set("Cache-Control", "public, max-age=60")
	w.Header().Set("ETag", etag)

	// Evaluate `If-None-Match` header. If client's cache matches server ETag, return 304 [4].
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified) // 304 Not Modified carries zero response body bytes [5]
		return
	}

	// 3. Delegate Pagination & Filtering to Store Layer [11]
	dishes, total, err := app.dishes.Page(q, page, limit)
	if err != nil {
		app.serverError(w, err)
		return
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))
	if pages == 0 {
		pages = 1
	}

	filtersMeta := map[string]any{}
	if q.Fasting != nil {
		filtersMeta["fasting"] = *q.Fasting
	}

	// 4. Respond with Envelope and Meta Block [11, 17]
	writeJSON(w, http.StatusOK, envelope{
		"data": dishes,
		"meta": map[string]any{
			"page":    page,
			"limit":   limit,
			"total":   total,
			"pages":   pages,
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
		w.Header().Set("Cache-Control", "public, max-age=60")
		writeJSON(w, http.StatusOK, envelope{"data": dish})
	}
}
