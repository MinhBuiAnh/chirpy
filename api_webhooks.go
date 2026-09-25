package main

import (
	"encoding/json"
	"net/http"

	"github.com/MinhBuiAnh/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.apiKey {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	type reqParams struct {
		Event string `json:"event"`
		Data struct{
			UserId string `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqParams{}

	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}
	
	parsedUserId, err := uuid.Parse(params.Data.UserId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	if err := cfg.db.UpgradeChirpyRedById(r.Context(), parsedUserId); err != nil {
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}