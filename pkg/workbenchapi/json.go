package workbenchapi

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code string, message string, details interface{}) {
	writeJSON(w, status, ErrorResponse{Error: ErrorPayload{Code: code, Message: message, Details: details}})
}
