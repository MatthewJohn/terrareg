package terrareg

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

// RespondJSON sends a JSON response with the given status code.
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal Server Error"}`))
		log.Error().Err(err).Msg("Failed to marshal JSON response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	w.WriteHeader(status)
	w.Write(response)
}

// RespondError sends an error response with the given status code and message.
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]string{"error": message})
}

// RespondInternalServerError sends a generic 500 Internal Server Error.
func RespondInternalServerError(w http.ResponseWriter, err error, msg string) {
	log.Error().Err(err).Msg(msg)
	RespondError(w, http.StatusInternalServerError, "Internal Server Error")
}
