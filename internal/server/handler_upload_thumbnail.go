package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/andycostintoma/tubely/internal/database"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
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
		return NewApiError(http.StatusInternalServerError, "Couldn't get video", err)
	}
	if videoMetadata.UserID != userID {
		return NewApiError(http.StatusUnauthorized, "You do not have permission to upload this video", err)
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20 // 10 mb
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		return NewApiError(http.StatusBadRequest, "Couldn't parse multipart form", err)
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		return NewApiError(http.StatusBadRequest, "Unable to parse form file", err)
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

	if mediaType != "image/jpeg" && mediaType != "image/png" {
		return NewApiError(http.StatusBadRequest, "Invalid content type", err)
	}

	thumbnailUrl := ""
	if cfg.thumbnailsStorage == "db" {
		rawBytes, err := io.ReadAll(file)
		if err != nil {
			return NewApiError(http.StatusInternalServerError, "Couldn't read file bytes", err)
		}
		encodedImg := base64.StdEncoding.EncodeToString(rawBytes)
		thumbnailUrl = fmt.Sprintf("data:%v;base64,%v", mediaType, encodedImg)

	} else if cfg.thumbnailsStorage == "fs" {
		fileExt := strings.Split(mediaType, "/")[1]
		randomBytes := make([]byte, 32)
		_, err = rand.Read(randomBytes)
		if err != nil {
			return NewApiError(http.StatusInternalServerError, "Couldn't generate random bytes", err)
		}
		randomString := base64.RawURLEncoding.EncodeToString(randomBytes)
		filePath := fmt.Sprintf("%v/%v.%v", cfg.assetsRoot, randomString, fileExt)
		newFile, err := os.Create(filePath)
		if err != nil {
			return NewApiError(http.StatusInternalServerError, "Couldn't create file", err)
		}
		defer newFile.Close()
		_, err = io.Copy(newFile, file)
		if err != nil {
			return NewApiError(http.StatusInternalServerError, "Couldn't copy file", err)
		}
		thumbnailUrl = fmt.Sprintf("%v:%v/%v", cfg.serverURL, cfg.port, filePath)
	}

	updatedVideo := database.Video{
		ID:                videoMetadata.ID,
		CreatedAt:         videoMetadata.CreatedAt,
		UpdatedAt:         time.Now(),
		ThumbnailURL:      &thumbnailUrl,
		VideoURL:          videoMetadata.VideoURL,
		CreateVideoParams: videoMetadata.CreateVideoParams,
	}

	err = cfg.db.UpdateVideo(updatedVideo)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't update video", err)
	}

	respondWithJSON(w, http.StatusOK, updatedVideo)
	return nil
}
