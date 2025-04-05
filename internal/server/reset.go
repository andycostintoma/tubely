package server

import (
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) error {
	if cfg.platform != "dev" {
		return NewApiError(http.StatusForbidden, "Reset is only allowed in dev environment.", nil)
	}

	err := cfg.db.Reset()
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't reset database", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Database reset to initial state"))
	return nil
}
