package msteams

import (
	"encoding/json"
	"net/http"
)

type OutgoingResponse struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// outputJSON outputs json responses
func outputJSON(w http.ResponseWriter, httpCode int, jsonStruct interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)

	json.NewEncoder(w).Encode(jsonStruct)
}
