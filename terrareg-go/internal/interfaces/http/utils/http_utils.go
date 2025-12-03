package httputils

import (
	"encoding/json"
	"net/http"
	"strings"
)

// decodeJSON parses JSON from request body
func DecodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

// sendJSONResponse sends a JSON response
func SendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// sendErrorResponse sends a JSON error response
func SendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	SendJSONResponse(w, statusCode, map[string]interface{}{
		"error": message,
	})
}

// extractBearerToken extracts Bearer token from Authorization header
func ExtractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Extract Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}