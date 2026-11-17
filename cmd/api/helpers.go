package main

import (
	"encoding/json"
	"net/http"
)

// envelope standardizes the top-level response container across all endpoints.
type envelope map[string]any

func writeJSON(w http.ResponseWriter, status int, data envelope) {
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, `{"error":{"message":"internal error"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, envelope{"error": map[string]string{"message": message}})
}

// serverError logs the detailed error and outputs a 500 response.
// IT IS THE ONLY FUNCTION IN THE CODEBASE AUTHORIZED TO WRITE StatusInternalServerError.
func (app *application) serverError(w http.ResponseWriter, err error) {
	app.logger.Println("SERVER ERROR:", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}

func (app *application) notFound(w http.ResponseWriter, msg string) {
	writeError(w, http.StatusNotFound, msg)
}

func (app *application) badRequest(w http.ResponseWriter, msg string) {
	writeError(w, http.StatusBadRequest, msg)
}