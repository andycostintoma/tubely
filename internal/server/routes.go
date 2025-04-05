package server

import "net/http"

func (cfg *apiConfig) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(cfg.filepathRoot)))
	mux.Handle("/app/", appHandler)

	assetsHandler := http.StripPrefix("/assets", http.FileServer(http.Dir(cfg.assetsRoot)))
	mux.Handle("/assets/", noCacheMiddleware(assetsHandler))

	mux.HandleFunc("POST /api/users", withApiError(cfg.handlerUsersCreate))
	mux.HandleFunc("POST /api/login", withApiError(cfg.handlerLogin))
	mux.HandleFunc("POST /api/refresh", withApiError(cfg.handlerRefresh))
	mux.HandleFunc("POST /api/revoke", withApiError(cfg.handlerRevoke))
	mux.HandleFunc("GET /api/videos/{videoID}", withApiError(cfg.handlerVideoGet))
	mux.HandleFunc("POST /admin/reset", withApiError(cfg.handlerReset))

	mux.HandleFunc("POST /api/videos", cfg.withAuth(cfg.handlerVideoMetaCreate))
	mux.HandleFunc("POST /api/thumbnail_upload/{videoID}", cfg.withAuth(cfg.handlerUploadThumbnail))
	mux.HandleFunc("POST /api/video_upload/{videoID}", cfg.withAuth(cfg.handlerUploadVideo))
	mux.HandleFunc("GET /api/videos", cfg.withAuth(cfg.handlerVideosRetrieve))
	mux.HandleFunc("DELETE /api/videos/{videoID}", cfg.withAuth(cfg.handlerVideoMetaDelete))

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
