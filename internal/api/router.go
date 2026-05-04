package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/cors"
	"github.com/trithemius/parq/internal/config"
	"github.com/trithemius/parq/internal/store"
)

type Server struct {
	store    store.RowStore
	cfg      *config.Config
	static   string
	basePath string
}

func NewRouter(s store.RowStore, cfg *config.Config, staticDir, corsOrigins, basePath string) http.Handler {
	srv := &Server{
		store:    s,
		cfg:      cfg,
		static:   staticDir,
		basePath: basePath,
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/schema", srv.handleGetSchema)
	mux.HandleFunc("GET /api/meta", srv.handleGetMeta)
	mux.HandleFunc("GET /api/stats", srv.handleGetStats)
	mux.HandleFunc("GET /api/rows", srv.handleGetRows)
	mux.HandleFunc("GET /api/rows/{idx}", srv.handleGetRow)
	mux.HandleFunc("PUT /api/rows/{idx}", srv.handleUpdateRow)
	mux.HandleFunc("GET /api/subdirs", srv.handleGetSubdirs)
	mux.HandleFunc("GET /api/thumbnail", srv.handleGetThumbnail)
	mux.HandleFunc("GET /api/file", srv.handleGetFile)
	mux.HandleFunc("GET /api/rows/{idx}/download", srv.handleDownload)
	mux.HandleFunc("POST /api/discover", srv.handleDiscover)

	// Static files
	mux.HandleFunc("/", srv.handleStatic)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{corsOrigins},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
	})

	return c.Handler(mux)
}

func (s *Server) handleGetSchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.cfg)
}

func (s *Server) handleGetMeta(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats, _ := s.store.Stats()
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"stats":  stats,
		"config": s.cfg,
	})
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.static, r.URL.Path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(s.static, "index.html"))
		return
	}
	http.FileServer(http.Dir(s.static)).ServeHTTP(w, r)
}
