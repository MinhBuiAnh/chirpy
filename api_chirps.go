package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/MinhBuiAnh/chirpy/internal/auth"
	"github.com/MinhBuiAnh/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type reqParams struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqParams{}

	w.Header().Set("Content-Type", "application/json")

	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	userId, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedBody := replaceBadWords(params.Body)
	dbParams := database.CreateChirpParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body: cleanedBody,
		UserID: userId,
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), dbParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	resBody := Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}
	respondWithJSON(w, http.StatusCreated, resBody)
}

func (cfg *apiConfig) handleGetChirpById(w http.ResponseWriter, r *http.Request) {
	pathVal := r.PathValue("chirpId")

	w.Header().Set("Content-Type", "application/json")

	chirpId, err := uuid.Parse(pathVal)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	chirp, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	resBody := Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}
	respondWithJSON(w, http.StatusOK, resBody)
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authorId := r.URL.Query().Get("author_id")
	sortingOrder := r.URL.Query().Get("sort")
	var chirps []database.Chirp

	if authorId == "" { // Get all chirps in the database
		var err error
		chirps, err = cfg.db.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, errorMessage)
			return
		}
	} else {
		parsedAuthorId, err := uuid.Parse(authorId)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, errorMessage)
			return
		}

		chirps, err = cfg.db.GetChirpsByUserId(r.Context(), parsedAuthorId)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, errorMessage)
			return
		}
	}

	var resBody []Chirp
	for _, chirp := range chirps {
		newChirp := Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		}
		resBody = append(resBody, newChirp)
	}

	if sortingOrder == "desc" { // Sort in descending order
		sort.SliceStable(resBody, func(i, j int) bool {
			return resBody[i].CreatedAt.After(resBody[j].CreatedAt)
		})
	} else { // Sort in descending order, default to this when no sort param found
		sort.SliceStable(resBody, func(i, j int) bool {
			return resBody[i].CreatedAt.Before(resBody[j].CreatedAt)
		})
	}

	respondWithJSON(w, http.StatusOK, resBody)
}

func (cfg *apiConfig) handleDeleteChirpById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	userId, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	pathVal := r.PathValue("chirpID")
	chirpId, err := uuid.Parse(pathVal)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	chirp, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	if userId != chirp.UserID {
		respondWithError(w, http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	if err := cfg.db.DeleteChirpById(r.Context(), chirpId); err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}