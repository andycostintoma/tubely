package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
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

func CreateTempFile(r io.Reader, extension string) (*os.File, error) {
	bytes, err := GenerateRandomBytes(16)
	if err != nil {
		return nil, err
	}
	filename := fmt.Sprintf("%s.%s", hex.EncodeToString(bytes), extension)

	temp, err := os.CreateTemp("", filename)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(temp, r)
	if err != nil {
		return nil, err
	}

	_, err = temp.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return temp, nil
}
