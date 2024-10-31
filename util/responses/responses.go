package responses

import (
	"encoding/json"
	"log"
	"net/http"
)

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
