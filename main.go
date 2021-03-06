package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// Response definition for response API
type Response struct {
	Message string `json:"message"`
}

func handler(w http.ResponseWriter, r *http.Request) {

	response := Response{}
	response.Message = "Hello world!"
	responseJSON, err := json.Marshal(response)

	if err != nil {
		log.Fatalf("Failed to parse json: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
