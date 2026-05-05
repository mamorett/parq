package api

import (
	"net/http"
	"strconv"

	"github.com/trithemius/parq/internal/thumbnail"
)

func (s *Server) handleGetThumbnail(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("parquet")
	if name == "" {
		names := s.store.StoreNames()
		if len(names) > 0 {
			name = names[0]
		}
	}

	store, cfg, err := s.getStoreAndConfig(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	idxStr := q.Get("idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "invalid index", http.StatusBadRequest)
		return
	}

	col := q.Get("col")
	if col == "" {
		col = cfg.Thumbnail.Column
	}

	row, err := store.Get(idx)
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

	w.Header().Set("Content-Type", "image/jpeg")
	err = thumbnail.Generate(path, cfg.Thumbnail.MaxSize, cfg.Thumbnail.Format, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
