package responses

import (
	"encoding/json"
	"log"
	"net/http"
)

func HttpErrorResponse(w http.ResponseWriter, message string, error string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	if message == "" {
		message = "An unexpected error occurred"
	}
	log.Println("Error: " + message)
	response := struct {
		Message   string `json:"message"`
		Error     string `json:"error"`
		ErrorCode int    `json:"errorCode"`
	}{
		Message:   message,
		Error:     error,
		ErrorCode: statusCode,
	}
	ResponseWithJSON(w, response, statusCode)
}

func ResponseWithJSON(w http.ResponseWriter, response interface{}, statusCode int) {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println("Error converting to JSON:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Println("Error sending response body:", err)
	}
}
