# Tubely

Tubely is a video management platform that allows users to upload, manage, and process videos and thumbnails. It provides a RESTful API for backend operations and a simple web interface for user interaction.

## Features

### User Authentication
- Sign up, log in, and manage sessions using JWT-based authentication.
- Refresh tokens for session management with token revocation support.

### Video Management
- Create video drafts with metadata (title, description).
- Upload videos and thumbnails.
- Process videos for optimized playback using `ffmpeg`.

### Storage Options
- Store videos and thumbnails in:
  - Local filesystem.
  - Database.
  - AWS S3 (with LocalStack support for local development).

### API Endpoints
- RESTful API for user management, video uploads, and metadata handling.
- Middleware for authentication and error handling.

### Frontend
- Web interface for user login, video draft creation, and video management.
- Built with HTML, CSS, and JavaScript.

### Backend
- Written in Go with a modular structure for database, server, and utility functions.
- SQLite database with auto-migration for tables.

### Local Development
- Docker Compose for running LocalStack to simulate AWS services.
- Makefile for building, running, testing, and managing the application.

### Video Processing
- Analyze video aspect ratios and categorize as landscape, portrait, or other.
- Process videos for fast start playback.

## Installation

### Prerequisites
- Go 1.23 or later
- Docker and Docker Compose
- `ffmpeg` and `ffprobe` installed on your system

### Steps
1. Clone the repository:
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Set up environment variables:
   - Copy `.env.example` to `.env` and update the values as needed.

4. Start LocalStack (optional for S3 simulation):
   ```bash
   make start-localstack
   ```

5. Build and run the application:
   ```bash
   make build
   make run
   ```

## Usage

### Web Interface
- Open the web interface in your browser at `http://localhost:<PORT>` (default: 8091).
- Log in or sign up to start managing videos.

### API Endpoints
- Use tools like Postman or `curl` to interact with the RESTful API.
- Refer to the `routes.go` file for available endpoints.

### Makefile Commands
- Build the application: `make build`
- Run the application: `make run`
- Test the application: `make test`
- Start LocalStack: `make start-localstack`
- Stop LocalStack: `make stop-localstack`

## Development

### Directory Structure
- `app/`: Frontend files (HTML, CSS, JavaScript).
- `cmd/`: Main entry point for the application.
- `internal/`: Backend logic, including authentication, database, server, and utilities.

### Testing
- Run tests using:
  ```bash
  make test
  ```