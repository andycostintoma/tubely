package server

import (
	"fmt"
	"github.com/andycostintoma/tubely/internal/database"
	"github.com/google/uuid"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
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

	fmt.Println("uploading video for video", videoID, "by user", userID)

	const maxMemory = 1 << 30 // 1 GB
	file, mediaType, err := readFormFile(r, "video", maxMemory)
	if err != nil {
		return err
	}
	defer file.Close()

	if mediaType != "video/mp4" {
		return NewApiError(http.StatusBadRequest, "Invalid content type", err)
	}

	storage := &S3Storage{
		Client:        cfg.s3Client,
		Bucket:        cfg.s3Bucket,
		Region:        cfg.s3Region,
		UseLocalstack: cfg.useLocalstack,
		LocalstackURL: cfg.localstackURL,
	}

	videoURL, err := storage.Save(r.Context(), file, mediaType)
	if err != nil {
		return NewInternalServerError(err)
	}

	updatedVideo := database.Video{
		ID:                videoMetadata.ID,
		CreatedAt:         videoMetadata.CreatedAt,
		UpdatedAt:         time.Now(),
		ThumbnailURL:      videoMetadata.ThumbnailURL,
		VideoURL:          &videoURL,
		CreateVideoParams: videoMetadata.CreateVideoParams,
	}

	err = cfg.db.UpdateVideo(updatedVideo)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't update video", err)
	}

	return nil
}
