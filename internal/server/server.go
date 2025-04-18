package server

import (
	"context"
	"fmt"
	"github.com/andycostintoma/tubely/internal/database"
	"github.com/andycostintoma/tubely/internal/utils"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"net/http"
	"os"
	"time"
)

type apiConfig struct {
	serverURL         string
	port              string
	platform          string
	db                database.Client
	jwtSecret         string
	filepathRoot      string
	assetsRoot        string
	thumbnailsStorage string
	useLocalstack     bool
	localstackURL     string
	s3Client          *s3.Client
	s3URLMode         string
	s3Bucket          string
	s3Region          string
	s3CfDistribution  string
}

func newApiConfig() (*apiConfig, error) {

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		return nil, fmt.Errorf("environment variable SERVER_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		return nil, fmt.Errorf("environment variable PORT is not set")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		return nil, fmt.Errorf("environment variable PLATFORM is not set")
	}

	pathToDB := os.Getenv("DB_PATH")
	if pathToDB == "" {
		return nil, fmt.Errorf("environment variable DB_PATH is not set")
	}

	db, err := database.NewClient(pathToDB)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}

	thumbnailStorage := os.Getenv("THUMBNAILS_STORAGE")
	if thumbnailStorage == "" {
		return nil, fmt.Errorf("environment variable THUMBNAILS_STORAGE is not set")
	}
	if thumbnailStorage != "db" && thumbnailStorage != "fs" {
		return nil, fmt.Errorf("THUMBNAILS_STORAGE %s is not allowed. Must be one of: db, fs", thumbnailStorage)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("environment variable JWT_SECRET is not set")
	}

	filepathRoot := os.Getenv("FILEPATH_ROOT")
	if filepathRoot == "" {
		return nil, fmt.Errorf("environment variable FILEPATH_ROOT is not set")
	}

	assetsRoot := os.Getenv("ASSETS_ROOT")
	if assetsRoot == "" {
		return nil, fmt.Errorf("environment variable ASSETS_ROOT is not set")
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		return nil, fmt.Errorf("environment variable S3_REGION is not set")
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		return nil, fmt.Errorf("environment variable S3_BUCKET is not set")
	}

	s3URLMode := os.Getenv("S3_URL_MODE")
	if s3URLMode == "" {
		return nil, fmt.Errorf("environment variable S3_URL_MODE is not set")
	}
	if s3URLMode != "localstack" && s3URLMode != "public" && s3URLMode != "presigned" && s3URLMode != "cloudfront" {
		return nil, fmt.Errorf("S3_URL_MODE %s is not allowed. Must be public, presigned or clodfront", s3URLMode)
	}

	useLocalstack := s3URLMode == "localstack"
	localstackURL := os.Getenv("LOCALSTACK_URL")
	if useLocalstack {
		if localstackURL == "" {
			return nil, fmt.Errorf("environment variable LOCALSTACK_URL is not set")
		}
	}

	s3CfDistribution := os.Getenv("S3_CF_DISTRO")
	if s3URLMode == "cloudfront" {
		if s3CfDistribution == "" {
			return nil, fmt.Errorf("environment variable S3_CF_DISTRO is not set")
		}
	}

	awsConfig, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(s3Region))

	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = useLocalstack
	})

	if err != nil {
		return nil, fmt.Errorf("could not load default aws config: %v", err)
	}

	cfg := apiConfig{
		serverURL:         serverURL,
		port:              port,
		platform:          platform,
		db:                db,
		jwtSecret:         jwtSecret,
		filepathRoot:      filepathRoot,
		assetsRoot:        assetsRoot,
		thumbnailsStorage: thumbnailStorage,
		localstackURL:     localstackURL,
		s3Client:          s3Client,
		s3URLMode:         s3URLMode,
		s3Bucket:          s3Bucket,
		s3Region:          s3Region,
		s3CfDistribution:  s3CfDistribution,
	}

	return &cfg, nil
}

func NewServer() (*http.Server, error) {

	cfg, err := newApiConfig()
	if err != nil {
		return nil, err
	}

	err = utils.EnsureDirExists(cfg.assetsRoot, 0755)
	if err != nil {
		return nil, err
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", cfg.port),
		Handler:      cfg.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Server listening on %v:%v", cfg.serverURL, cfg.port)

	return server, nil
}
