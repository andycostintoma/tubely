package server

import (
	"errors"
	"fmt"
	"github.com/andycostintoma/tubely/internal/auth"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type ApiError struct {
	Code    int
	Message string
	Err     error
}

func NewApiError(code int, message string, err error) *ApiError {
	return &ApiError{Code: code, Message: message, Err: err}
}

func (a *ApiError) Error() string {
	if a.Err != nil {
		return fmt.Sprintf("%d %s: %v", a.Code, a.Message, a.Err)
	}
	return fmt.Sprintf("%d %s", a.Code, a.Message)
}

type ApiErrorHandlerFunc func(w http.ResponseWriter, r *http.Request) error

type errorResponse struct {
	Error string `json:"error"`
}

func withApiError(handler ApiErrorHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Recover from panic if one occurs
		defer func() {
			if rec := recover(); rec != nil {
				// Log the panic and respond with a 500
				log.Printf("HTTP %s %s -> 500: Panic: %v", r.Method, r.URL.Path, r)
				respondWithJSON(w, http.StatusInternalServerError, errorResponse{
					Error: "Internal Server Error",
				})
			}
		}()

		err := handler(w, r)
		if err == nil {
			return
		}

		var apiErr *ApiError
		if errors.As(err, &apiErr) {
			log.Printf("HTTP %s %s -> %d: %v", r.Method, r.URL.Path, apiErr.Code, apiErr.Err)
			respondWithJSON(w, apiErr.Code, errorResponse{
				Error: apiErr.Message,
			})
			return
		}

		// Fallback for unexpected errors
		log.Printf("HTTP %s %s -> 500: %v", r.Method, r.URL.Path, err)
		respondWithJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "Internal Server Error",
		})
		return
	}
}

type AuthenticatedHandlerFunc func(http.ResponseWriter, *http.Request, uuid.UUID) error

func (cfg *apiConfig) withAuth(handler AuthenticatedHandlerFunc) http.HandlerFunc {
	return withApiError(func(w http.ResponseWriter, r *http.Request) error {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			return NewApiError(http.StatusUnauthorized, "Unauthorized: Missing or invalid token", err)
		}

		userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
		if err != nil {
			return NewApiError(http.StatusUnauthorized, "Unauthorized: Invalid token", err)
		}

		return handler(w, r, userID)
	})
}
