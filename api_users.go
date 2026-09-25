package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MinhBuiAnh/chirpy/internal/auth"
	"github.com/MinhBuiAnh/chirpy/internal/database"
	"github.com/google/uuid"
)

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

	passwordHash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	dbParams := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email: params.Email,
		HashedPassword: sql.NullString{
			String: passwordHash,
			Valid: true,
		},
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
		IsChirpyRed: user.IsChirpyRed.Bool,
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

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	refreshTokenString := auth.MakeRefreshToken()
	revokeTime := sql.NullTime{
		Valid: false,
	}
	rtParams := database.CreateRefreshTokenParams{
		Token: refreshTokenString,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		ExpiresAt: time.Now().Add(3600 * time.Hour),
		RevokedAt: revokeTime,
	}
	if err := cfg.db.CreateRefreshToken(r.Context(), rtParams); err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	resBody := User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		Token: token,
		RefreshToken: refreshTokenString,
		IsChirpyRed: user.IsChirpyRed.Bool,
	}
	respondWithJSON(w, http.StatusOK, resBody)
}

func (cfg *apiConfig) handleUpdateEmailPassword(w http.ResponseWriter, r *http.Request) {
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

	type reqParams struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqParams{}

	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	updatedPasswordHash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	dbParams := database.UpdateUserParams{
		ID: userId,
		Email: params.Email,
		HashedPassword: sql.NullString{
			String: updatedPasswordHash,
			Valid: true,
		},
		UpdatedAt: time.Now(),
	}
	userUpdated, err := cfg.db.UpdateUser(r.Context(), dbParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	resBody := User{
		ID: userId,
		CreatedAt: userUpdated.CreatedAt,
		UpdatedAt: userUpdated.UpdatedAt,
		Email: userUpdated.Email,
		Token: tokenString,
		IsChirpyRed: userUpdated.IsChirpyRed.Bool,
	}
	respondWithJSON(w, http.StatusOK, resBody)
}