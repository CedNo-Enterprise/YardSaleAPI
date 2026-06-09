package interfaces

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func Decode(w http.ResponseWriter, decoder *json.Decoder, v any) {
	err := decoder.Decode(v)
	if err != nil {
		slog.Error("Error parsing request body", "err", err)
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}
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

func ValidateContentType(w http.ResponseWriter, r *http.Request, t string) {
	contentType := r.Header.Get("Content-Type")
	if contentType != t {
		http.Error(w, "invalid content type", http.StatusUnsupportedMediaType)
		return
	}
}
