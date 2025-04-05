package server

import "net/http"

func (cfg *apiConfig) RegisterRoutes() http.Handler {
	mux := NewCustomServeMux(cfg.jwtSecret)
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(cfg.filepathRoot)))
	mux.Handle("/app/", appHandler)

	assetsHandler := http.StripPrefix("/assets", http.FileServer(http.Dir(cfg.assetsRoot)))
	mux.Handle("/assets/", noCacheMiddleware(assetsHandler))

	mux.HandleApiError("POST /api/users", cfg.handlerUsersCreate)
	mux.HandleApiError("POST /api/login", cfg.handlerLogin)
	mux.HandleApiError("POST /api/refresh", cfg.handlerRefresh)
	mux.HandleApiError("POST /api/revoke", cfg.handlerRevoke)

	mux.HandleAuthenticated("POST /api/videos", cfg.handlerVideoMetaCreate)
	mux.HandleAuthenticated("POST /api/thumbnail_upload/{videoID}", cfg.handlerUploadThumbnail)
	mux.HandleAuthenticated("POST /api/video_upload/{videoID}", cfg.handlerUploadVideo)
	mux.HandleAuthenticated("GET /api/videos", cfg.handlerVideosRetrieve)
	mux.HandleApiError("GET /api/videos/{videoID}", cfg.handlerVideoGet)
	mux.HandleAuthenticated("DELETE /api/videos/{videoID}", cfg.handlerVideoMetaDelete)

	mux.HandleApiError("POST /admin/reset", cfg.handlerReset)

	// Wrap the mux with CORS middleware
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}
