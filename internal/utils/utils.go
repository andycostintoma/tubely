package utils

import (
	"crypto/rand"
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
