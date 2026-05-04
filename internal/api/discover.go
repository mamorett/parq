package api

import (
	"encoding/json"
	"net/http"

	"github.com/trithemius/parq/internal/config"
)

func (s *Server) handleDiscover(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ParquetFile string `json:"parquet_file"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cfg, err := config.Discover(body.ParquetFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}
