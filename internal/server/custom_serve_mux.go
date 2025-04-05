package server

import (
	"errors"
	"fmt"
	"github.com/andycostintoma/tubely/internal/auth"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type CustomServeMux struct {
	mux       *http.ServeMux
	jwtSecret string
}

func NewCustomServeMux(jwtSecret string) *CustomServeMux {
	return &CustomServeMux{
		mux:       http.NewServeMux(),
		jwtSecret: jwtSecret,
	}
}

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
type AuthenticatedHandlerFunc func(http.ResponseWriter, *http.Request, uuid.UUID) error

func (c *CustomServeMux) HandleApiError(pattern string, handler ApiErrorHandlerFunc) {
	c.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err == nil {
			return
		}

		var apiErr *ApiError
		type errorResponse struct {
			Error string `json:"error"`
		}
		if errors.As(err, &apiErr) {
			log.Printf("HTTP %s %s -> %d: %v", r.Method, r.URL.Path, apiErr.Code, apiErr.Err)
			respondWithJSON(w, apiErr.Code, errorResponse{
				Error: apiErr.Message,
			})
			return
		}

		// Fallback for unexpected errors
		log.Printf("HTTP %s %s -> 500: %v", r.Method, r.URL.Path, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
}

func (c *CustomServeMux) HandleAuthenticated(pattern string, handler AuthenticatedHandlerFunc) {
	c.HandleApiError(pattern, func(w http.ResponseWriter, r *http.Request) error {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			return NewApiError(http.StatusUnauthorized, "Unauthorized: Missing or invalid token", err)
		}

		userID, err := auth.ValidateJWT(token, c.jwtSecret)
		if err != nil {
			return NewApiError(http.StatusUnauthorized, "Unauthorized: Invalid token", err)
		}

		return handler(w, r, userID)
	})
}

func (c *CustomServeMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	c.mux.HandleFunc(pattern, handler)
}

func (c *CustomServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

func (c *CustomServeMux) Handle(s string, handler http.Handler) {
	c.mux.Handle(s, handler)
}
