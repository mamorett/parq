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
	store    *store.MultiStore
	cfg      *config.Config // single config for backward compat (first/default parquet)
	multiCfg *config.MultiConfig
	static   string
	basePath string
}

func NewRouter(ms *store.MultiStore, mc *config.MultiConfig, staticDir, corsOrigins, basePath string) http.Handler {
	srv := &Server{
		store:    ms,
		multiCfg: mc,
		static:   staticDir,
		basePath: basePath,
	}

	// Set default config to first parquet (for backward compat)
	if len(mc.Parquets) > 0 {
		if cfg, err := mc.Parquets[0].Resolve(); err == nil {
			srv.cfg = cfg
		}
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/schema", srv.handleGetSchema)
	mux.HandleFunc("GET /api/meta", srv.handleGetMeta)
	mux.HandleFunc("GET /api/stats", srv.handleGetStats)
	mux.HandleFunc("GET /api/rows", srv.handleGetRows)
	mux.HandleFunc("GET /api/rows/{idx}", srv.handleGetRow)
	mux.HandleFunc("PUT /api/rows/{idx}", srv.handleUpdateRow)
	mux.HandleFunc("DELETE /api/rows/{idx}", srv.handleDeleteRow)
	mux.HandleFunc("GET /api/subdirs", srv.handleGetSubdirs)
	mux.HandleFunc("GET /api/thumbnail", srv.handleGetThumbnail)
	mux.HandleFunc("GET /api/file", srv.handleGetFile)
	mux.HandleFunc("GET /api/rows/{idx}/download", srv.handleDownload)
	mux.HandleFunc("POST /api/discover", srv.handleDiscover)
	mux.HandleFunc("GET /api/parquets", srv.handleGetParquets)

	// Static files
	mux.HandleFunc("/", srv.handleStatic)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{corsOrigins},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
	})

	return c.Handler(mux)
}

// getStoreAndConfig extracts the store and config for a given parquet name
func (s *Server) getStoreAndConfig(name string) (*store.MemoryStore, *config.Config, error) {
	store, err := s.store.StoreFor(name)
	if err != nil {
		return nil, nil, err
	}
	cfg, err := s.store.GetConfig(name)
	if err != nil {
		return nil, nil, err
	}
	return store, cfg, nil
}

func (s *Server) handleGetSchema(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("parquet")
	if name == "" {
		// Return first/default config
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.cfg)
		return
	}

	cfg, err := s.store.GetConfig(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

func (s *Server) handleGetMeta(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("parquet")
	if name == "" {
		name = s.store.StoreNames()[0]
	}

	store, err := s.store.StoreFor(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	cfg, err := s.store.GetConfig(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	stats, _ := store.Stats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"stats":  stats,
		"config": cfg,
	})
}

func (s *Server) handleGetParquets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"parquets": s.store.StoreNames(),
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
