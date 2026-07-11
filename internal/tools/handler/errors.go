package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func WriteJSON(
	w http.ResponseWriter,
	status int,
	v any,
) {

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WriteError(
	w http.ResponseWriter,
	status int,
	message, err string,
) {
	log.Println("http err", message, err)
	WriteJSON(
		w,
		status,
		ErrorResponse{
			Status: status,
			Message: message,
			Error: err,
		},
	)
}
