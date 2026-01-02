package httpserver

import (
	"io/fs"
	"net/http"
	"strings"

	webui "github.com/koeppj/go-app-manager/web/ui"
)

func (s *Server) registerRoutes() {
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.tlsMiddleware)
	s.router.Use(s.cidr.Middleware)

	api := s.router.PathPrefix("/api").Subrouter()
	api.Use(s.authMiddleware())
	api.HandleFunc("/health", s.handleHealth).Methods(http.MethodGet)
	api.HandleFunc("/programs", s.handlePrograms).Methods(http.MethodGet)
	api.HandleFunc("/programs/{name}", s.handleProgram).Methods(http.MethodGet)
	api.HandleFunc("/programs/{name}/start", s.handleProgramStart).Methods(http.MethodPost)
	api.HandleFunc("/programs/{name}/stop", s.handleProgramStop).Methods(http.MethodPost)
	api.HandleFunc("/programs/{name}/restart", s.handleProgramRestart).Methods(http.MethodPost)

	// UI assets: custom SPA handler to avoid redirects and directory canonicalization.
	subFS, err := fs.Sub(webui.Static, "static")
	if err != nil {
		return
	}

	spaHandler := func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if pathFileExists(subFS, path) {
			http.ServeFileFS(w, r, subFS, path)
			return
		}
		http.ServeFileFS(w, r, subFS, "index.html")
	}

	s.router.PathPrefix("/").HandlerFunc(spaHandler)
}

// pathFileExists checks if a file exists in the embedded FS.
func pathFileExists(fsys fs.FS, name string) bool {
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		return false
	}
	return true
}
