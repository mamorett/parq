package api

import (
	"net/http"
	"strconv"

	"github.com/trithemius/parq/internal/thumbnail"
)

func (s *Server) handleGetThumbnail(w http.ResponseWriter, r *http.Request) {
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

	// In Phase 2, we would apply path remapping here if needed
	// For now, assume path is valid or relative to CWD

	w.Header().Set("Content-Type", "image/jpeg")
	err = thumbnail.Generate(path, s.cfg.Thumbnail.MaxSize, s.cfg.Thumbnail.Format, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
