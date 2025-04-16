package server

import (
	"encoding/json"
	"github.com/andycostintoma/tubely/internal/database"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerVideoMetaCreate(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	type parameters struct {
		database.CreateVideoParams
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't decode parameters", err)
	}
	params.UserID = userID

	video, err := cfg.db.CreateVideo(params.CreateVideoParams)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't create video", err)
	}

	respondWithJSON(w, http.StatusCreated, video)
	return nil
}

func (cfg *apiConfig) handlerVideoMetaDelete(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		return NewApiError(http.StatusBadRequest, "Invalid ID", err)
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		return NewApiError(http.StatusNotFound, "Couldn't get video", err)
	}
	if video.UserID != userID {
		return NewApiError(http.StatusForbidden, "You can't delete this video", err)
	}

	err = cfg.db.DeleteVideo(videoID)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't delete video", err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (cfg *apiConfig) handlerVideoGet(w http.ResponseWriter, r *http.Request) error {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		return NewApiError(http.StatusBadRequest, "Invalid video ID", err)
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		return NewApiError(http.StatusNotFound, "Couldn't get video", err)
	}

	if cfg.s3URLMode == "signed" {
		video, err = cfg.dbVideoToSignedVideo(r.Context(), video)
		if err != nil {
			return NewInternalServerError(err)
		}
	}

	respondWithJSON(w, http.StatusOK, video)
	return nil
}

func (cfg *apiConfig) handlerVideosRetrieve(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	videos, err := cfg.db.GetVideos(userID)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, "Couldn't retrieve videos", err)
	}

	if cfg.s3URLMode == "signed" {
		signedVideos := make([]database.Video, 0, len(videos))
		for _, video := range videos {
			signedVideo, err := cfg.dbVideoToSignedVideo(r.Context(), video)
			if err != nil {
				return NewInternalServerError(err)
			}
			signedVideos = append(signedVideos, signedVideo)
		}
		videos = signedVideos
	}

	respondWithJSON(w, http.StatusOK, videos)
	return nil
}
