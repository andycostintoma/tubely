package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func GetVideoAspectRatio(filePath string) (string, error) {
	command := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var out bytes.Buffer
	command.Stdout = &out
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("failed to run ffprobe: %v", err)
	}

	type ffprobeOutput struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	var probeData ffprobeOutput
	if err := json.Unmarshal(out.Bytes(), &probeData); err != nil {
		return "", fmt.Errorf("failed to unmarshal ffprobe output: %w", err)
	}

	if len(probeData.Streams) == 0 || probeData.Streams[0].Width == 0 || probeData.Streams[0].Height == 0 {
		return "", errors.New("no valid video stream found")
	}

	width := probeData.Streams[0].Width
	height := probeData.Streams[0].Height
	ratio := float64(width) / float64(height)
	const epsilon = 0.01

	switch {
	case floatsEqual(ratio, 16.0/9.0, epsilon):
		return "16:9", nil
	case floatsEqual(ratio, 9.0/16.0, epsilon):
		return "9:16", nil
	default:
		return "other", nil
	}
}

func ProcessVideoForFastStart(inputFilePath string) (string, error) {
	processedFilePath := fmt.Sprintf("%s.processing", inputFilePath)

	cmd := exec.Command("ffmpeg", "-i", inputFilePath, "-movflags", "faststart", "-codec", "copy", "-f", "mp4", processedFilePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error processing video: %s, %v", stderr.String(), err)
	}

	fileInfo, err := os.Stat(processedFilePath)
	if err != nil {
		return "", fmt.Errorf("could not stat processed file: %v", err)
	}
	if fileInfo.Size() == 0 {
		return "", fmt.Errorf("processed file is empty")
	}

	return processedFilePath, nil
}
