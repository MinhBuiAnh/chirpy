package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/MinhBuiAnh/chirpy/internal/auth"
	"github.com/MinhBuiAnh/chirpy/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
}

// FUNCTIONS FOR METRICS OPERATION
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	content := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
	w.Write([]byte(content))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	if err := cfg.db.DeleteAllUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type reqParams struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqParams{}

	w.Header().Set("Content-Type", "application/json")

	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	passwordParam := sql.NullString{
		String: hashedPassword,
		Valid: true,
	}
	dbParams := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email: params.Email,
		HashedPassword: passwordParam,
	}
	user, err := cfg.db.CreateUser(r.Context(), dbParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	resBody := User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}
	respondWithJSON(w, http.StatusCreated, resBody)
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type reqParams struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqParams{}

	w.Header().Set("Content-Type", "application/json")

	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword.String)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	resBody := User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}
	respondWithJSON(w, http.StatusOK, resBody)
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type reqParams struct {
		Body string `json:"body"`
		UserID string `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqParams{}

	w.Header().Set("Content-Type", "application/json")

	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedBody := replaceBadWords(params.Body)
	userIdParam, err := uuid.Parse(params.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}
	dbParams := database.CreateChirpParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body: cleanedBody,
		UserID: userIdParam,
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

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
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

	respondWithJSON(w, http.StatusOK, resBody)
}