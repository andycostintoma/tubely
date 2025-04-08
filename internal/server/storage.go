package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/andycostintoma/tubely/internal/utils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"os"
	"strings"
)

type Storage interface {
	Save(ctx context.Context, r io.Reader, contentType string) (string, error)
}

type FSStorage struct {
	AssetsRoot string
	ServerURL  string
	Port       string
}

func (fs *FSStorage) Save(_ context.Context, r io.Reader, mediaType string) (string, error) {
	fileExt := strings.Split(mediaType, "/")[1]
	randomBytes, err := utils.GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}
	randomString := base64.RawURLEncoding.EncodeToString(randomBytes)

	filePath := fmt.Sprintf("%v/%v.%v", fs.AssetsRoot, randomString, fileExt)

	newFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, r)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v:%v/%v", fs.ServerURL, fs.Port, filePath), nil
}

type DBStorage struct{}

func (db *DBStorage) Save(_ context.Context, r io.Reader, mediaType string) (string, error) {
	rawBytes, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(rawBytes)
	return fmt.Sprintf("data:%v;base64,%v", mediaType, encoded), nil
}

type S3Storage struct {
	Client        *s3.Client
	Bucket        string
	Filename      string
	Region        string
	UseLocalstack bool
	LocalstackURL string
}

func (s *S3Storage) Save(ctx context.Context, r io.Reader, mediaType string) (string, error) {

	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.Bucket,
		Key:         &s.Filename,
		Body:        r,
		ContentType: &mediaType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	if s.UseLocalstack {
		return fmt.Sprintf("%s/%s/%s", s.LocalstackURL, s.Bucket, s.Filename), nil
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.Bucket, s.Region, s.Filename), nil
}
