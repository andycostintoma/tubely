package server

import (
	"fmt"
	"github.com/andycostintoma/tubely/internal/database"
	"mime"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	videoIDString := r.PathValue("videoID")

	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		return NewApiError(http.StatusBadRequest, "Invalid ID", err)
	}

	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		return NewInternalServerError(err)
	}
	if videoMetadata.UserID != userID {
		return NewApiError(http.StatusUnauthorized, "You do not have permission to upload this video", err)
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20 // 10 mb
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		return NewApiError(http.StatusBadRequest, "Error parsing multipart form", err)
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		return NewApiError(http.StatusBadRequest, "Error parsing form file", err)
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		return NewApiError(http.StatusBadRequest, "Error getting content type", nil)
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return NewApiError(http.StatusBadRequest, "Error parsing media type", err)
	}

	if mediaType != "image/jpeg" && mediaType != "image/png" {
		return NewApiError(http.StatusBadRequest, "Invalid content type", err)
	}

	var storage Storage
	switch cfg.thumbnailsStorage {
	case "db":
		storage = &FSStorage{AssetsRoot: cfg.assetsRoot, ServerURL: cfg.serverURL, Port: cfg.port}
	case "fs":
		storage = &DBStorage{}
	default:
		return NewInternalServerError(fmt.Errorf("invalid storage type: %s", cfg.thumbnailsStorage))
	}

	thumbnailURL, err := storage.Save(r.Context(), file, mediaType)
	if err != nil {
		return NewInternalServerError(err)
	}

	updatedVideo := database.Video{
		ID:                videoMetadata.ID,
		CreatedAt:         videoMetadata.CreatedAt,
		UpdatedAt:         time.Now(),
		ThumbnailURL:      &thumbnailURL,
		VideoURL:          videoMetadata.VideoURL,
		CreateVideoParams: videoMetadata.CreateVideoParams,
	}

	err = cfg.db.UpdateVideo(updatedVideo)
	if err != nil {
		return NewInternalServerError(err)
	}

	respondWithJSON(w, http.StatusOK, updatedVideo)
	return nil
}
