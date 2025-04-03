package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/andycostintoma/tubely/internal/auth"
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

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")

	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondAndLog(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondAndLog(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondAndLog(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondAndLog(w, http.StatusInternalServerError, "Couldn't get video", err)
		return
	}
	if videoMetadata.UserID != userID {
		respondAndLog(w, http.StatusUnauthorized, "You do not have permission to upload this video", err)
		return
	}

	fmt.Println("uploading video for video", videoID, "by user", userID)

	const maxMemory = 1 << 30 // 1 GB
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondAndLog(w, http.StatusInternalServerError, "Couldn't parse multipart form", err)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondAndLog(w, http.StatusInternalServerError, "Couldn't get file", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		respondAndLog(w, http.StatusBadRequest, "No content type specified", err)
		return
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondAndLog(w, http.StatusBadRequest, "Invalid content type", err)
		return
	}

	if mediaType != "video/mp4" {
		respondAndLog(w, http.StatusBadRequest, "Invalid media type", err)
		return
	}

	temp, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondAndLog(w, http.StatusInternalServerError, "Couldn't create temp file", err)
		return
	}
	defer os.Remove(temp.Name())
	defer temp.Close()

	_, err = io.Copy(temp, file)
	if err != nil {
		respondAndLog(w, http.StatusInternalServerError, "Couldn't copy to temp file", err)
		return
	}

	_, err = temp.Seek(0, io.SeekStart)
	if err != nil {
		respondAndLog(w, http.StatusInternalServerError, "Couldn't reset temp file", err)
		return
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
		respondAndLog(w, http.StatusInternalServerError, "Couldn't upload video", err)
		return
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
		respondAndLog(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}

}
