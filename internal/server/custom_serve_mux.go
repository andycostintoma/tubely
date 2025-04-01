package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondAndLog(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}

type AppResponse struct {
	StatusCode int
	Message    string
	Err        error
}

type CustomServerMux struct {
	mux *http.ServeMux
}

type CustomHandlerFunc func(http.ResponseWriter, *http.Request) AppResponse

func NewCustomServerMux() *CustomServerMux {
	return &CustomServerMux{
		mux: http.NewServeMux(),
	}
}

func (c *CustomServerMux) Handle(pattern string, handler CustomHandlerFunc) {
	c.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		appResponse := handler(w, r)
		if appResponse.Err != nil {
			respondAndLog(w, appResponse.StatusCode, appResponse.Message, appResponse.Err)
		} else {
			respondWithJSON(w, appResponse.StatusCode, appResponse.Message)
		}
	})
}
