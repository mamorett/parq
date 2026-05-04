package api

import (
	"encoding/json"
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

	row, err := s.store.Get(idx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"row_%d.json\"", idx))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(row.Columns); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
