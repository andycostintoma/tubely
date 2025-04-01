package server

import (
	"fmt"
	"github.com/andycostintoma/tubely/internal/database"
	"log"
	"net/http"
	"os"
	"time"
)

type apiConfig struct {
	db               database.Client
	jwtSecret        string
	platform         string
	serverURL        string
	filepathRoot     string
	assetsRoot       string
	s3Bucket         string
	s3Region         string
	s3CfDistribution string
	port             string
}

func NewServer() (*http.Server, error) {

	pathToDB := os.Getenv("DB_PATH")
	if pathToDB == "" {
		return nil, fmt.Errorf("environment variable DB_PATH is not set")
	}

	db, err := database.NewClient(pathToDB)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}

	serverUrl := os.Getenv("SERVER_URL")
	if serverUrl == "" {
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

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		return nil, fmt.Errorf("environment variable S3_BUCKET is not set")
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		return nil, fmt.Errorf("environment variable S3_REGION is not set")
	}

	s3CfDistribution := os.Getenv("S3_CF_DISTRO")
	if s3CfDistribution == "" {
		return nil, fmt.Errorf("environment variable S3_CF_DISTRO is not set")
	}

	cfg := apiConfig{
		db:               db,
		jwtSecret:        jwtSecret,
		platform:         platform,
		filepathRoot:     filepathRoot,
		assetsRoot:       assetsRoot,
		s3Bucket:         s3Bucket,
		s3Region:         s3Region,
		s3CfDistribution: s3CfDistribution,
		port:             port,
	}

	err = cfg.ensureAssetsDir()
	if err != nil {
		log.Fatalf("Couldn't create assets directory: %v", err)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", cfg.port),
		Handler:      cfg.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Server listening on %v:%v", serverUrl, port)

	return server, nil
}
