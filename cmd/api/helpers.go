package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"unicode"
)

var errInvalidID = errors.New("invalid id format")

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

// readID validates path params (length <= 32, alphanumeric + hyphen).
func readID(r *http.Request) (string, error) {
	id := r.PathValue("id")
	if id == "" || len(id) > 32 {
		return "", errInvalidID
	}
	for _, c := range id {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' {
			return "", errInvalidID
		}
	}
	return id, nil
}

func readString(qs url.Values, key, def string) string {
	if v := qs.Get(key); v != "" {
		return v
	}
	return def
}

func readBool(qs url.Values, key string) (value, ok bool) {
	v := qs.Get(key)
	if v == "" {
		return false, false
	}
	b, err := strconv.ParseBool(v)
	return b, err == nil
}

func readSort(qs url.Values, allowed []string, def string) (string, bool) {
	v := readString(qs, "sort", def)
	return v, slices.Contains(allowed, v)
}