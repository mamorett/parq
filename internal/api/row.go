package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) handleGetRow(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(row)
}

func (s *Server) handleUpdateRow(w http.ResponseWriter, r *http.Request) {
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

	var cols map[string]any
	if err := json.NewDecoder(r.Body).Decode(&cols); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := store.Update(idx, cols); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *Server) handleDeleteRow(w http.ResponseWriter, r *http.Request) {
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

	if err := store.Delete(idx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
