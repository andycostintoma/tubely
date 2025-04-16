package server

import (
	"fmt"
	"github.com/andycostintoma/tubely/internal/database"
	"github.com/andycostintoma/tubely/internal/utils"
	"github.com/google/uuid"
	"mime"
	"net/http"
	"os"
	"path/filepath"
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
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		return NewApiError(http.StatusBadRequest, "Error parsing multipart form", err)
	}

	file, header, err := r.FormFile("video")
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

	if mediaType != "video/mp4" {
		return NewApiError(http.StatusBadRequest, "Invalid content type", err)
	}

	temp, err := utils.CreateTempFile(file, "mp4")
	if err != nil {
		return NewInternalServerError(err)
	}
	defer os.Remove(temp.Name())
	defer temp.Close()

	ratio, err := utils.GetVideoAspectRatio(temp.Name())
	if err != nil {
		return NewInternalServerError(err)
	}

	filename := filepath.Base(temp.Name())
	switch ratio {
	case "16:9":
		filename = fmt.Sprintf("landscape/%v", filename)
	case "9:16":
		filename = fmt.Sprintf("portrait/%v", filename)
	default:
		filename = fmt.Sprintf("other/%v", filename)
	}

	processedFileName, err := utils.ProcessVideoForFastStart(temp.Name())
	if err != nil {
		return NewInternalServerError(err)
	}
	processedFile, err := os.Open(processedFileName)
	if err != nil {
		return NewInternalServerError(err)
	}
	defer processedFile.Close()
	defer os.Remove(processedFile.Name())

	storage := &S3Storage{
		Client:        cfg.s3Client,
		Region:        cfg.s3Region,
		Bucket:        cfg.s3Bucket,
		Key:           filename,
		URLMode:       cfg.s3URLMode,
		LocalstackURL: cfg.localstackURL,
		CloudFrontURL: cfg.s3CfDistribution,
	}

	videoURL, err := storage.Save(r.Context(), processedFile, mediaType)
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
