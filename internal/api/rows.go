package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/trithemius/parq/internal/store"
)

func (s *Server) handleGetRows(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(q.Get("size"))
	if size < 1 {
		size = s.cfg.Pagination.DefaultPageSize
	}

	filter := store.Filter{
		Search:    q.Get("search"),
		SearchCol: q.Get("search_col"),
		Exact:     make(map[string]string),
		Subdirs:   q["subdir"],
	}

	sort := store.Sort{
		Column: q.Get("sort"),
		Order:  q.Get("order"),
	}
	if sort.Column == "" {
		sort.Column = s.cfg.DefaultSort.Column
		sort.Order = s.cfg.DefaultSort.Order
	}

	rows, total, err := s.store.Query(filter, sort, store.Pagination{Page: page, Size: size})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"total": total,
		"page":  page,
		"size":  size,
		"rows":  rows,
	})
}
