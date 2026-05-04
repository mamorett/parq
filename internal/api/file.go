package api

import (
	"net/http"
	"strconv"
)

func (s *Server) handleGetFile(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	idxStr := q.Get("idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "invalid index", http.StatusBadRequest)
		return
	}

	col := q.Get("col")
	if col == "" {
		col = s.cfg.Thumbnail.Column
	}

	row, err := s.store.Get(idx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	pathVal, ok := row.Columns[col]
	if !ok {
		http.Error(w, "column not found", http.StatusBadRequest)
		return
	}

	path, ok := pathVal.(string)
	if !ok {
		http.Error(w, "invalid path column", http.StatusBadRequest)
		return
	}

	// Apply remapping if exists
	finalPath := path
	// In a real scenario we'd use the rewriter from store
	// For now we'll serve the file directly
	
	http.ServeFile(w, r, finalPath)
}
