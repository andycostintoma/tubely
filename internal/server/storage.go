package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/andycostintoma/tubely/internal/database"
	"github.com/andycostintoma/tubely/internal/utils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"os"
	"strings"
	"time"
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
	Region        string
	Bucket        string
	Key           string
	URLMode       string
	LocalstackURL string
	CloudFrontURL string
}

func (s *S3Storage) Save(ctx context.Context, r io.Reader, mediaType string) (string, error) {

	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.Bucket,
		Key:         &s.Key,
		Body:        r,
		ContentType: &mediaType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	switch s.URLMode {
	case "localstack":
		return fmt.Sprintf("%s/%s/%s", s.LocalstackURL, s.Bucket, s.Key), nil
	case "public":
		return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.Bucket, s.Region, s.Key), nil
	case "presigned":
		return fmt.Sprintf("%s,%s", s.Bucket, s.Key), nil
	case "cloudfront":
		return fmt.Sprintf("https://%s/%s", s.CloudFrontURL, s.Key), nil
	default:
		return "", errors.New("unsupported URL mode")
	}
}

func generatePreSignedURL(context context.Context, s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	preSignClient := s3.NewPresignClient(s3Client)

	req, err := preSignClient.PresignGetObject(context, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expireTime))

	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(context context.Context, video database.Video) (database.Video, error) {
	videoURL := video.VideoURL
	parts := strings.Split(*videoURL, ",")
	if len(parts) != 2 {
		return database.Video{}, errors.New("invalid video URL")
	}
	bucket := parts[0]
	key := parts[1]
	signedUrl, err := generatePreSignedURL(context, cfg.s3Client, bucket, key, 15*time.Minute)
	if err != nil {
		return database.Video{}, err
	}
	return database.Video{
		ID:                video.ID,
		CreatedAt:         video.CreatedAt,
		UpdatedAt:         time.Now(),
		ThumbnailURL:      video.ThumbnailURL,
		VideoURL:          &signedUrl,
		CreateVideoParams: video.CreateVideoParams,
	}, nil
}
