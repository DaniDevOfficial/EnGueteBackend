package error

import (
	"log"
	"net/http"
	"wishtournament/util/responses"
)

func HttpResponse(w http.ResponseWriter, message string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	if message == "" {
		message = "An unexpected error occurred"
	}
	log.Println("Error: " + message)
	response := struct {
		Message   string `json:"message"`
		ErrorCode int    `json:"errorCode"`
	}{
		Message:   message,
		ErrorCode: statusCode,
	}
	responses.ResponseWithJSON(w, response, statusCode)
}
