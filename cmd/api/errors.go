package main

import (
	"errors"
	"net/http"

	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/menu"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/order"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/validator"
)

// failedValidation formats field-level validation failures into a 422 JSON envelope [3].
func (app *application) failedValidation(w http.ResponseWriter, fields map[string]string) {
	writeJSON(w, http.StatusUnprocessableEntity, envelope{
		"error": map[string]any{
			"message": "the request was not valid",
			"fields":  fields,
		},
	})
}

// handleError acts as the single exit door for handler errors.
// It maps internal domain errors to appropriate HTTP status codes [10, 11].
func (app *application) handleError(w http.ResponseWriter, err error) {
	var vErr *validator.Error
	switch {
	case errors.As(err, &vErr):
		// 422 Unprocessable Entity for validation failures [10]
		app.failedValidation(w, vErr.Fields)
	case errors.Is(err, menu.ErrNotFound), errors.Is(err, order.ErrNotFound):
		// 404 Not Found for missing resources [10]
		app.notFound(w, "resource not found")
	case errors.Is(err, order.ErrIllegalTransition):
		// 409 Conflict for invalid state machine steps [10]
		app.conflict(w, err.Error())
	default:
		// 500 Internal Server Error for unmapped server bugs (logged in full) [10, 22]
		app.serverError(w, err)
	}
}