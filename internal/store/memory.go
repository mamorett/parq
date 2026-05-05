package store

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/trithemius/parq/internal/config"
	"github.com/trithemius/parq/internal/image"
	"github.com/trithemius/parq/internal/parquet"
	"github.com/trithemius/parq/internal/pathrewrite"
	"github.com/trithemius/parq/internal/stats"
	libparquet "github.com/parquet-go/parquet-go"
)

type MemoryStore struct {
	mu        sync.RWMutex
	cfg       *config.Config
	data      []map[string]any
	stats     stats.Stats
	rewriters map[string]*pathrewrite.Rewriter
	metaCache map[int]*image.ImageMeta
}

func NewMemoryStore(cfg *config.Config) (*MemoryStore, error) {
	s := &MemoryStore{
		cfg:       cfg,
		rewriters: make(map[string]*pathrewrite.Rewriter),
		metaCache: make(map[int]*image.ImageMeta),
	}

	if cfg != nil {
		for _, col := range cfg.Columns {
			if col.Type == "path" && len(col.Remap) > 0 {
				rw, err := pathrewrite.New(col.Remap)
				if err == nil {
					s.rewriters[col.Name] = rw
				}
			}
		}

		if cfg.ParquetFile != "" {
			if err := s.Reload(); err != nil {
				return nil, err
			}
		}
	}

	return s, nil
}

func (s *MemoryStore) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := parquet.ReadAll(s.cfg.ParquetFile)
	if err != nil {
		return err
	}
	s.data = data
	s.metaCache = make(map[int]*image.ImageMeta)

	info, err := os.Stat(s.cfg.ParquetFile)
	if err != nil {
		return err
	}

	s.stats = stats.Compute(data, s.cfg)
	s.stats.FileSize = info.Size()

	return nil
}

func (s *MemoryStore) Query(f Filter, sortOpt Sort, p Pagination) ([]Row, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]Row, 0)
	for i, r := range s.data {
		if s.matches(r, f) {
			row := Row{Index: i, Columns: r}
			if meta, ok := s.metaCache[i]; ok {
				row.ImageMeta = meta
			}
			filtered = append(filtered, row)
		}
	}

	total := len(filtered)

	// Sorting
	if sortOpt.Column != "" {
		sort.Slice(filtered, func(i, j int) bool {
			valI := filtered[i].Columns[sortOpt.Column]
			valJ := filtered[j].Columns[sortOpt.Column]
			res := compare(valI, valJ)
			if sortOpt.Order == "desc" {
				return res > 0
			}
			return res < 0
		})
	}

	// Pagination
	start := (p.Page - 1) * p.Size
	if start >= total {
		return []Row{}, total, nil
	}
	end := start + p.Size
	if end > total {
		end = total
	}

	result := filtered[start:end]

	// Post-process: probe dimensions if needed for the current page
	for i := range result {
		idx := result[i].Index
		if result[i].ImageMeta != nil {
			continue
		}
		// Probe only if configured
		for _, col := range s.cfg.Columns {
			if col.ProbeDimensions && col.Type == "path" {
				if pathVal, ok := result[i].Columns[col.Name].(string); ok {
					// Apply rewrite if exists
					finalPath := pathVal
					if rw, ok := s.rewriters[col.Name]; ok {
						finalPath = rw.Rewrite(pathVal)
					}
					meta, err := image.Probe(finalPath)
					if err == nil {
						// Update cache safely
						s.mu.RUnlock()
						s.mu.Lock()
						s.metaCache[idx] = meta
						s.mu.Unlock()
						s.mu.RLock()
						result[i].ImageMeta = meta
					} else {
						slog.Warn("Failed to probe image", "path", finalPath, "error", err)
					}
				}
				break
			}
		}
	}

	return result, total, nil
}

func (s *MemoryStore) Get(idx int) (Row, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if idx < 0 || idx >= len(s.data) {
		return Row{}, fmt.Errorf("index out of bounds")
	}
	row := Row{Index: idx, Columns: s.data[idx]}
	if meta, ok := s.metaCache[idx]; ok {
		row.ImageMeta = meta
	}
	return row, nil
}

func (s *MemoryStore) Update(idx int, cols map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if idx < 0 || idx >= len(s.data) {
		return fmt.Errorf("index out of bounds")
	}

	for k, v := range cols {
		s.data[idx][k] = v
	}

	// Persist back to parquet
	f, err := os.Open(s.cfg.ParquetFile)
	if err != nil {
		return err
	}
	info, _ := f.Stat()
	pf, err := libparquet.OpenFile(f, info.Size())
	if err != nil {
		f.Close()
		return err
	}
	schema := pf.Schema()
	f.Close()

	return parquet.WriteAll(s.cfg.ParquetFile, s.data, schema)
}

func (s *MemoryStore) Stats() (stats.Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats, nil
}

func (s *MemoryStore) Subdirs(colName string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	unique := make(map[string]struct{})
	for _, r := range s.data {
		val, ok := r[colName]
		if !ok {
			continue
		}
		path, ok := val.(string)
		if !ok {
			continue
		}
		parts := strings.Split(path, "/")
		for _, p := range parts {
			if p != "" && !strings.Contains(p, ".") {
				unique[p] = struct{}{}
			}
		}
	}

	res := make([]string, 0, len(unique))
	for k := range unique {
		res = append(res, k)
	}
	sort.Strings(res)
	return res, nil
}

func (s *MemoryStore) Close() error {
	return nil
}

// GetConfig returns the config for this store
func (s *MemoryStore) GetConfig() *config.Config {
	return s.cfg
}

func (s *MemoryStore) matches(r map[string]any, f Filter) bool {
	if f.Search != "" {
		match := false
		searchLower := strings.ToLower(f.Search)
		if f.SearchCol != "" {
			if val, ok := r[f.SearchCol]; ok {
				if strings.Contains(strings.ToLower(fmt.Sprintf("%v", val)), searchLower) {
					match = true
				}
			}
		} else {
			for _, col := range s.cfg.Columns {
				if !col.Searchable {
					continue
				}
				if val, ok := r[col.Name]; ok {
					if strings.Contains(strings.ToLower(fmt.Sprintf("%v", val)), searchLower) {
						match = true
						break
					}
				}
			}
		}
		if !match {
			return false
		}
	}

	for col, target := range f.Exact {
		if val, ok := r[col]; ok {
			if fmt.Sprintf("%v", val) != target {
				return false
			}
		} else {
			return false
		}
	}

	for _, subdir := range f.Subdirs {
		found := false
		for _, col := range s.cfg.Columns {
			if col.Type != "path" {
				continue
			}
			if val, ok := r[col.Name]; ok {
				path := fmt.Sprintf("%v", val)
				if strings.Contains(path, "/"+subdir+"/") || strings.HasPrefix(path, subdir+"/") || strings.HasSuffix(path, "/"+subdir) {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func compare(a, b any) int {
	strA := strings.ToLower(fmt.Sprintf("%v", a))
	strB := strings.ToLower(fmt.Sprintf("%v", b))
	if strA == strB {
		return 0
	}
	if strA < strB {
		return -1
	}
	return 1
}
