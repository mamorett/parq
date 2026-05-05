package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
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

	idxStr := r.PathValue("idx")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "invalid index", http.StatusBadRequest)
		return
	}

	row, err := store.Get(idx)
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
