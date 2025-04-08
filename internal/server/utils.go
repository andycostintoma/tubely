package server

import (
	"encoding/json"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
)

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

func readFormFile(r *http.Request, field string, maxMemory int64) (file multipart.File, mediaType string, apiError *ApiError) {
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		return nil, "", NewApiError(http.StatusBadRequest, "Error parsing multipart form", err)
	}

	file, header, err := r.FormFile(field)
	if err != nil {
		return nil, "", NewApiError(http.StatusBadRequest, "Error retrieving the file", err)
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		return nil, "", NewApiError(http.StatusBadRequest, "No content type specified", nil)
	}

	mediaType, _, err = mime.ParseMediaType(contentType)
	if err != nil {
		return nil, "", NewApiError(http.StatusBadRequest, "Error parsing media type", err)
	}

	return file, mediaType, nil
}
