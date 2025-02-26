package main

import (
	"encoding/json"
	"http_server/internal/auth"
	"net/http"

	"github.com/google/uuid"
)

func (c *apiConfig) upgradeUser(w http.ResponseWriter, r *http.Request) {
	type requestStruct struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	apikey, err := auth.GetAPIKey(r.Header)
	if err != nil || apikey != c.polka_key {
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
		return
	}

	var req requestStruct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Could not decode request", http.StatusUnprocessableEntity)
		return
	}
	defer r.Body.Close()
	if req.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	err = c.queries.UpgradeUser(r.Context(), req.Data.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
	}
	w.WriteHeader(http.StatusNoContent)
}
