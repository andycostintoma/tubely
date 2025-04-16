# Build the application
all: build test

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Download files
download-samples:
	@mkdir -p samples
	@urls="\
		https://storage.googleapis.com/qvault-webapp-dynamic-assets/course_assets/boots-image-horizontal.png \
		https://storage.googleapis.com/qvault-webapp-dynamic-assets/course_assets/boots-image-vertical.png \
		https://storage.googleapis.com/qvault-webapp-dynamic-assets/course_assets/boots-video-horizontal.mp4 \
		https://storage.googleapis.com/qvault-webapp-dynamic-assets/course_assets/boots-video-vertical.mp4 \
		https://storage.googleapis.com/qvault-webapp-dynamic-assets/course_assets/is-bootdev-for-you.pdf"; \
	for url in $$urls; do \
		file_name=$$(basename "$$url"); \
		echo "Downloading $$file_name..."; \
		curl -sSfL -o "samples/$$file_name" "$$url"; \
	done

start-localstack:
	@echo "Starting LocalStack..."
	docker compose up -d

stop-localstack:
	@echo "Stopping LocalStack..."
	docker compose down

configure-localstack:
	@echo "Configuring AWS CLI profile for LocalStack"
	aws configure set aws_access_key_id test
	aws configure set aws_secret_access_key test
	aws configure set region us-east-1
	aws configure set output json
	aws configure set endpoint_url http://localhost:4566

	aws iam create-user --user-name test-user
	aws iam create-group --group-name managers
	aws iam attach-group-policy --group-name managers --policy-arn arn:aws:iam::aws:policy/AdministratorAccess
	aws iam add-user-to-group --user-name test-user --group-name managers
	aws iam create-access-key --user-name test-user

	aws s3 mb s3://tubely-12345
	aws s3api put-bucket-policy --bucket tubely-12345 --policy file://s3-policy-localstack.json

.PHONY: all build run test clean watch download-samples start-localstack stop-localstack configure-localstack
