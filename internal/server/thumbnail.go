package server

import "github.com/google/uuid"

type thumbnail struct {
	data      []byte
	mediaType string
}

var videoThumbnails = map[uuid.UUID]thumbnail{}

