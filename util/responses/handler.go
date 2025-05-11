package responses

import (
	"encoding/json"
	"enguete/util/frontendErrors"
	"log"
	"net/http"
)

func HttpErrorResponse(w http.ResponseWriter, statusCode int, error string, message string) {
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

func GenericInternalServerError(w http.ResponseWriter) {
	HttpErrorResponse(w, http.StatusInternalServerError, frontendErrors.InternalServerError, "An unexpected error occurred")
}
func GenericBadRequestError(w http.ResponseWriter) {
	HttpErrorResponse(w, http.StatusBadRequest, frontendErrors.BadRequestError, "The request was invalid")
}
func GenericUnauthorizedError(w http.ResponseWriter) {
	HttpErrorResponse(w, http.StatusUnauthorized, frontendErrors.UnauthorizedError, "You are not authorized to perform this action")
}
func GenericForbiddenError(w http.ResponseWriter) {
	HttpErrorResponse(w, http.StatusForbidden, frontendErrors.ForbiddenError, "You do not have permission to perform this action")
}

func GenericNotFoundError(w http.ResponseWriter) {
	HttpErrorResponse(w, http.StatusNotFound, frontendErrors.NotFoundError, "The requested resource was not found")
}

func GenericNotAllowedToPerformActionError(w http.ResponseWriter) {
	HttpErrorResponse(w, http.StatusForbidden, frontendErrors.NotAllowedToPerformActionError, "You are not allowed to perform this action")
}

func GenericGroupDoesNotExistError(w http.ResponseWriter) {
	HttpErrorResponse(w, http.StatusNotFound, frontendErrors.GroupDoesNotExistError, "The group does not exist")
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
