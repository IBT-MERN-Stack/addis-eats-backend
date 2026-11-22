package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/delivery"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/menu"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/order"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/validator"
)

type createOrderRequest struct {
	Area    string   `json:"area"`
	Phone   string   `json:"phone"`
	DishIDs []string `json:"dish_ids"`
}

// Normalise canonicalises user input prior to validation judgment [8, 9].
func (in *createOrderRequest) Normalise() {
	in.Area = strings.TrimSpace(in.Area)

	// Clean telephone formatting (+251, spaces, hyphens) [8]
	p := strings.NewReplacer(" ", "", "-", "").Replace(in.Phone)
	p = strings.TrimPrefix(p, "+251")
	if strings.HasPrefix(p, "251") {
		p = strings.TrimPrefix(p, "251")
	}
	if !strings.HasPrefix(p, "0") && len(p) == 9 {
		p = "0" + p // Convert 911223344 -> 0911223344 [8]
	}
	in.Phone = p
}

// Validate executes all field checks and records failures into the Validator [20, 23].
func (in createOrderRequest) Validate(v *validator.Validator) {
	v.Check(in.Phone != "", "phone", "must be provided")
	v.Check(validator.Matches(in.Phone, validator.PhoneRx), "phone", "must be a valid TeleBirr number")
	v.Check(in.Area != "", "area", "must be provided")
	v.Check(validator.In(in.Area, delivery.ServedAreas...), "area", "must be a served delivery area")
	v.Check(len(in.DishIDs) > 0, "dish_ids", "must not be empty")
	v.Check(len(in.DishIDs) <= 50, "dish_ids", "must not contain more than 50 items")
	v.Check(validator.Unique(in.DishIDs), "dish_ids", "must not contain duplicates")
}

type patchOrderRequest struct {
	Status *order.Status `json:"status"`
}

func (app *application) createOrder(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	// 1. PARSE — Decode JSON structure (400 Bad Request on failure) [2]
	var in createOrderRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		app.badRequest(w, "malformed JSON body or unknown field")
		return
	}

	// 2. NORMALISE & VALIDATE — Canonicalise and validate input values [2, 8]
	in.Normalise()
	v := validator.New()
	in.Validate(v)
	if !v.Valid() {
		app.failedValidation(w, v.Errors) // 422 Unprocessable Entity [2, 23]
		return
	}

	// Check dish existence in store (maps missing dish ID to 422 error envelope) [4, 18, 21]
	items := make([]menu.Dish, 0, len(in.DishIDs))
	vErr := &validator.Error{Fields: make(map[string]string)}
	for _, id := range in.DishIDs {
		d, err := app.dishes.ByID(id)
		if err != nil {
			vErr.Fields["dish_ids"] = fmt.Sprintf("dish %s does not exist", id)
			app.handleError(w, vErr)
			return
		}
		items = append(items, d)
	}

	// 3. ACT — Execute domain placement [2]
	o := order.New(in.Area, in.Phone, items)
	_, err := o.ComputeTotal()
	if err != nil {
		app.handleError(w, err)
		return
	}
	if err := app.orders.Create(&o); err != nil {
		app.handleError(w, err)
		return
	}

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
	if err != nil {
		app.handleError(w, err)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, envelope{"data": o})
}

func (app *application) patchOrder(w http.ResponseWriter, r *http.Request) {
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
		app.failedValidation(w, map[string]string{"status": "must be provided"})
		return
	}

	updatedOrder, err := app.orders.UpdateStatus(id, *in.Status)
	if err != nil {
		app.handleError(w, err) // Maps ErrIllegalTransition to 409 Conflict [10]
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, envelope{"data": updatedOrder})
}