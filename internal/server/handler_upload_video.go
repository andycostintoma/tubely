package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/andycostintoma/tubely/internal/database"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
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
		return NewApiError(http.StatusInternalServerError, "Couldn't get video", err)
	}
	if videoMetadata.UserID != userID {
		return NewApiError(http.StatusUnauthorized, "You do not have permission to upload this video", err)
	}

	fmt.Println("uploading video for video", videoID, "by user", userID)

	const maxMemory = 1 << 30 // 1 GB
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't parse multipart form", err)
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't get file", err)
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		return NewApiError(http.StatusBadRequest, "No content type specified", err)
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return NewApiError(http.StatusBadRequest, "Invalid content type", err)
	}

	if mediaType != "video/mp4" {
		return NewApiError(http.StatusBadRequest, "Invalid content type", err)
	}

	temp, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't create temp file", err)
	}
	defer os.Remove(temp.Name())
	defer temp.Close()

	_, err = io.Copy(temp, file)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't copy to temp file", err)
	}

	_, err = temp.Seek(0, io.SeekStart)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't reset temp file", err)
	}

	fileExt := strings.Split(mediaType, "/")[1]
	bytes := make([]byte, 16)
	_, err = rand.Read(bytes)
	if err != nil {

	}
	filename := hex.EncodeToString(bytes) + "." + fileExt

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &filename,
		Body:        temp,
		ContentType: &mediaType,
	})

	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't upload video", err)
	}

	var videoURL string
	if cfg.useLocalstack {
		videoURL = fmt.Sprintf("%v/%v/%v", cfg.localstackURL, cfg.s3Bucket, filename)
	} else {
		videoURL = fmt.Sprintf("https://%v.s3.%v.amazonaws.com/%v", cfg.s3Bucket, cfg.s3Region, filename)
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
