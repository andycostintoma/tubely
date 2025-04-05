package server

import (
	"github.com/andycostintoma/tubely/internal/auth"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) error {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return NewApiError(http.StatusUnauthorized, "Unauthorized: Missing or invalid token", err)
	}

	user, err := cfg.db.GetUserByRefreshToken(refreshToken)
	if err != nil {
		return NewApiError(http.StatusUnauthorized, "Couldn't get user for refresh token", err)
	}

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,
		time.Hour,
	)
	if err != nil {
		return NewApiError(http.StatusUnauthorized, "Couldn't validate token", err)
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
	return nil
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) error {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return NewApiError(http.StatusUnauthorized, "Couldn't find token", err)
	}

	err = cfg.db.RevokeRefreshToken(refreshToken)
	if err != nil {
		return NewApiError(http.StatusUnauthorized, "Couldn't revoke token", err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
