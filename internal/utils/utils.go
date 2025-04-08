package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
)

func EnsureDirExists(dirPath string, permissions os.FileMode) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.Mkdir(dirPath, permissions)
	}
	return nil
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func floatsEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) <= epsilon
}

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
