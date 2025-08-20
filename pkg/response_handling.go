package pkg

import (
	"encoding/json"
	"log"
	"net/http"
)

func SendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	log.Printf("sending answer in json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error writing in json: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendError(w http.ResponseWriter, status int, message string) {
	SendJSON(w, status, ErrorResponse{Error: message})
}

func SendValidationError(w http.ResponseWriter, status int, errors map[string]string) {
	SendJSON(w, status, ValidationErrorResponse{Errors: errors})
}
