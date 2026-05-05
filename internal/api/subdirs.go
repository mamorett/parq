package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleGetSubdirs(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("parquet")
	if name == "" {
		names := s.store.StoreNames()
		if len(names) > 0 {
			name = names[0]
		}
	}

	store, err := s.store.StoreFor(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	col := r.URL.Query().Get("col")
	if col == "" {
		http.Error(w, "missing col parameter", http.StatusBadRequest)
		return
	}

	subdirs, err := store.Subdirs(col)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"column":  col,
		"subdirs": subdirs,
	})
}
