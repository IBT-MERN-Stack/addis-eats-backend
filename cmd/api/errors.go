package main

import (
	"errors"
	"net/http"

	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/menu"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/order"
	"github.com/IBT-MERN-STACK/addis-eats-backend/internal/validator"
)

func writeErrorWithReqID(w http.ResponseWriter, r *http.Request, status int, message string) {
	writeJSON(w, status, envelope{
		"error": map[string]any{
			"message":    message,
			"request_id": requestIDFrom(r), // Includes correlation ID in error responses [12]
		},
	})
}

func (app *application) failedValidation(w http.ResponseWriter, r *http.Request, fields map[string]string) {
	writeJSON(w, http.StatusUnprocessableEntity, envelope{
		"error": map[string]any{
			"message":    "the request was not valid",
			"fields":     fields,
			"request_id": requestIDFrom(r),
		},
	})
}

func (app *application) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var vErr *validator.Error
	switch {
	case errors.As(err, &vErr):
		app.failedValidation(w, r, vErr.Fields)
	case errors.Is(err, menu.ErrNotFound), errors.Is(err, order.ErrNotFound):
		writeErrorWithReqID(w, r, http.StatusNotFound, "resource not found")
	case errors.Is(err, order.ErrIllegalTransition):
		writeErrorWithReqID(w, r, http.StatusConflict, err.Error())
	default:
		app.logger.Println("SERVER ERROR:", err)
		writeErrorWithReqID(w, r, http.StatusInternalServerError, "internal error")
	}
}