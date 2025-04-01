package server

import (
	"encoding/base64"
	"fmt"
	"github.com/andycostintoma/tubely/internal/auth"
	"github.com/andycostintoma/tubely/internal/database"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20 // 10 mb
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondAndLog(w, http.StatusBadRequest, "Couldn't parse multipart form", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondAndLog(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")
	if mediaType == "" {
		respondAndLog(w, http.StatusBadRequest, "No content type specified", err)
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

	thumbnailUrl := ""
	if cfg.thumbnailsStorage == "db" {
		rawBytes, err := io.ReadAll(file)
		if err != nil {
			respondAndLog(w, http.StatusInternalServerError, "Couldn't read file bytes", err)
			return
		}
		encodedImg := base64.StdEncoding.EncodeToString(rawBytes)
		thumbnailUrl = fmt.Sprintf("data:%v;base64,%v", mediaType, encodedImg)

	} else if cfg.thumbnailsStorage == "fs" {
		fileExt := strings.Split(mediaType, "/")[1]
		filePath := fmt.Sprintf("%v/%v.%v", cfg.assetsRoot, videoID, fileExt)
		newFile, err := os.Create(filePath)
		if err != nil {
			respondAndLog(w, http.StatusInternalServerError, "Couldn't create file", err)
			return
		}
		defer newFile.Close()
		_, err = io.Copy(newFile, file)
		if err != nil {
			respondAndLog(w, http.StatusInternalServerError, "Couldn't copy file", err)
			return
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
		respondAndLog(w, http.StatusInternalServerError, "Couldn't update thumbnail", err)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedVideo)
}
