package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/MinhBuiAnh/chirpy/internal/auth"
	"github.com/MinhBuiAnh/chirpy/internal/database"
)

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), refreshTokenString)
	if err != nil || time.Now().After(refreshToken.ExpiresAt) || refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	token, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	type returnVals struct {
		Token string `json:"token"`
	}
	respondWithJSON(w, http.StatusOK, returnVals{ Token: token })
}

func (cfg *apiConfig) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}
	
	params := database.UpdateRefreshTokenParams{
		Token: refreshTokenString,
		RevokedAt: sql.NullTime{
			Time: time.Now(),
			Valid: true,
		},
	}
	if err := cfg.db.UpdateRefreshToken(r.Context(), params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errorMessage)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}