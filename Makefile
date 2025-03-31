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

.PHONY: all build run test clean watch download-samples
