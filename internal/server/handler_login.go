package server

import (
	"encoding/json"
	"github.com/andycostintoma/tubely/internal/auth"
	"github.com/andycostintoma/tubely/internal/database"

	"net/http"
	"time"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) error {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		database.User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		return NewApiError(http.StatusBadRequest, "Couldn't parse request body", err)
	}

	user, err := cfg.db.GetUserByEmail(params.Email)
	if err != nil {
		return NewApiError(http.StatusUnauthorized, "Incorrect email or password", err)
	}

	err = auth.CheckPasswordHash(params.Password, user.Password)
	if err != nil {
		return NewApiError(http.StatusUnauthorized, "Incorrect email or password", err)
	}

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,
		time.Hour*24*30,
	)

	if err != nil {
		return NewInternalServerError(err)
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		return NewInternalServerError(err)
	}

	_, err = cfg.db.CreateRefreshToken(database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		return NewInternalServerError(err)
	}

	respondWithJSON(w, http.StatusOK, response{
		User:         user,
		Token:        accessToken,
		RefreshToken: refreshToken,
	})

	return nil
}
