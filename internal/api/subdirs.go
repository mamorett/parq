package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleGetSubdirs(w http.ResponseWriter, r *http.Request) {
	col := r.URL.Query().Get("col")
	if col == "" {
		http.Error(w, "missing col parameter", http.StatusBadRequest)
		return
	}

	subdirs, err := s.store.Subdirs(col)
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
