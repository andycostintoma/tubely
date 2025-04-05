package server

import (
	"encoding/json"
	"github.com/andycostintoma/tubely/internal/auth"
	"github.com/andycostintoma/tubely/internal/database"
	"net/http"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) error {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't decode parameters", err)
	}

	if params.Password == "" || params.Email == "" {
		return NewApiError(http.StatusBadRequest, "Email and password are required", nil)
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't hash password", err)
	}

	user, err := cfg.db.CreateUser(database.CreateUserParams{
		Email:    params.Email,
		Password: hashedPassword,
	})
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't create user", err)
	}

	respondWithJSON(w, http.StatusCreated, user)
	return nil
}
