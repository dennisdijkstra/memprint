package main

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "applicatioin/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
