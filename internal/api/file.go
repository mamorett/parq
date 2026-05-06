package api

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/trithemius/parq/internal/pathrewrite"
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

	// Apply path rewriting if remap rules are configured for this column
	cfg, err := s.store.GetConfig(name)
	if err == nil {
		for _, colDef := range cfg.Columns {
			if colDef.Name == col && len(colDef.Remap) > 0 {
				rewriter, err := pathrewrite.New(colDef.Remap)
				if err == nil {
					path = rewriter.Rewrite(path)
				}
				break
			}
		}
	}

	// Check for download mode
	dl := q.Get("dl")
	if dl == "1" || dl == "true" {
		basename := filepath.Base(path)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+basename+"\"")
	}

	http.ServeFile(w, r, path)
}
