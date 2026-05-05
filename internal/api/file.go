package api

import (
	"net/http"
	"strconv"
)

func (s *Server) handleGetFile(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("parquet")
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

	idxStr := q.Get("idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "invalid index", http.StatusBadRequest)
		return
	}

	col := q.Get("col")
	if col == "" {
		// Get from config
		cfg, err := s.store.GetConfig(name)
		if err != nil || cfg.Thumbnail.Column == "" {
			http.Error(w, "no default column configured", http.StatusBadRequest)
			return
		}
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

	http.ServeFile(w, r, path)
}
