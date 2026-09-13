package interfaces

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

const maxRequestBodyBytes = 1048576

// DecodeBody validates the content type and decodes the JSON body, reporting
// whether the handler should continue. It writes the error response itself, so
// a caller that ignores the result would carry on past a rejected request and
// write a second status — always guard on the return value.
func DecodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "invalid content type", http.StatusUnsupportedMediaType)
		return false
	}

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		slog.Error("Error parsing request body", "err", err)
		http.Error(w, "bad request body", http.StatusBadRequest)
		return false
	}

	return true
}

func Encode(w http.ResponseWriter, v any) {
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		slog.Error("Error encoding response", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func Marshal(w http.ResponseWriter, v any) {
	_, err := json.Marshal(v)
	if err != nil {
		slog.Error("Error marshalling %T", v)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func WriteResponse(w http.ResponseWriter, response any, status int, contentType string) {
	// Headers must be set before WriteHeader: once the status is written, later
	// header changes are silently discarded and Go sniffs the type instead.
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	Encode(w, response)
}
