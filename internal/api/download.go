package api

import (
	"fmt"
	"net/http"
	"strconv"
)

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	idxStr := r.PathValue("idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "invalid index", http.StatusBadRequest)
		return
	}

	col := r.URL.Query().Get("col")
	if col == "" {
		http.Error(w, "missing col parameter", http.StatusBadRequest)
		return
	}

	row, err := s.store.Get(idx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	val, ok := row.Columns[col]
	if !ok {
		http.Error(w, "column not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_%d.txt\"", col, idx))
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprintf(w, "%v", val)
}
