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
	mu         sync.RWMutex
	cfg        *config.Config
	data       []map[string]any
	rowIDs     []int
	stats      stats.Stats
	rewriters  map[string]*pathrewrite.Rewriter
	metaCache  map[int]*image.ImageMeta
	schema     *libparquet.Schema
	skipReload bool
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

	if s.skipReload {
		slog.Info("Skipping watcher reload (self-write)", "file", s.cfg.ParquetFile)
		s.skipReload = false
		return nil
	}

	data, err := parquet.ReadAll(s.cfg.ParquetFile)
	if err != nil {
		return err
	}
	slog.Info("Reloaded parquet data", "file", s.cfg.ParquetFile, "rows", len(data))
	s.data = data

	s.rowIDs = make([]int, len(data))
	for i := range s.rowIDs {
		s.rowIDs[i] = i
	}
	s.metaCache = make(map[int]*image.ImageMeta)

	f, err := os.Open(s.cfg.ParquetFile)
	if err != nil {
		return fmt.Errorf("open parquet for schema: %w", err)
	}
	info, _ := f.Stat()
	pf, err := libparquet.OpenFile(f, info.Size())
	if err != nil {
		f.Close()
		return fmt.Errorf("read parquet schema: %w", err)
	}
	s.schema = pf.Schema()
	f.Close()

	fileInfo, err := os.Stat(s.cfg.ParquetFile)
	if err != nil {
		return err
	}

	s.stats = stats.Compute(data, s.cfg)
	s.stats.FileSize = fileInfo.Size()

	return nil
}

func (s *MemoryStore) Query(f Filter, sortOpt Sort, p Pagination) ([]Row, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]Row, 0)
	for i, r := range s.data {
		if s.matches(r, f) {
			rowID := s.rowIDs[i]
			row := Row{Index: rowID, Columns: r}
			if meta, ok := s.metaCache[rowID]; ok {
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

func (s *MemoryStore) findRowIdx(rowID int) (int, error) {
	for i, id := range s.rowIDs {
		if id == rowID {
			return i, nil
		}
	}
	return -1, fmt.Errorf("row %d not found", rowID)
}

func (s *MemoryStore) Get(rowID int) (Row, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idx, err := s.findRowIdx(rowID)
	if err != nil {
		return Row{}, err
	}
	row := Row{Index: rowID, Columns: s.data[idx]}
	if meta, ok := s.metaCache[rowID]; ok {
		row.ImageMeta = meta
	}
	return row, nil
}

func (s *MemoryStore) Update(rowID int, cols map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx, err := s.findRowIdx(rowID)
	if err != nil {
		return err
	}

	for k, v := range cols {
		s.data[idx][k] = v
	}

	if s.schema == nil {
		slog.Error("Cannot persist: parquet schema not loaded")
		return fmt.Errorf("parquet schema not loaded")
	}

	slog.Info("Persisting update to parquet", "file", s.cfg.ParquetFile, "rowID", rowID, "totalRows", len(s.data))
	if err := parquet.WriteAll(s.cfg.ParquetFile, s.data, s.schema); err != nil {
		slog.Error("Failed to persist parquet", "error", err)
		return fmt.Errorf("persist parquet: %w", err)
	}
	s.skipReload = true
	slog.Info("Parquet persisted successfully")
	return nil
}

func (s *MemoryStore) Delete(rowID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx, err := s.findRowIdx(rowID)
	if err != nil {
		return err
	}

	s.data = append(s.data[:idx], s.data[idx+1:]...)
	s.rowIDs = append(s.rowIDs[:idx], s.rowIDs[idx+1:]...)
	delete(s.metaCache, rowID)

	s.stats = stats.Compute(s.data, s.cfg)
	fileInfo, err := os.Stat(s.cfg.ParquetFile)
	if err == nil {
		s.stats.FileSize = fileInfo.Size()
	}

	if s.schema == nil {
		slog.Error("Cannot persist: parquet schema not loaded")
		return fmt.Errorf("parquet schema not loaded")
	}

	slog.Info("Persisting delete to parquet", "file", s.cfg.ParquetFile, "rowID", rowID, "totalRows", len(s.data))
	if err := parquet.WriteAll(s.cfg.ParquetFile, s.data, s.schema); err != nil {
		slog.Error("Failed to persist parquet after delete", "error", err)
		return fmt.Errorf("persist parquet: %w", err)
	}
	s.skipReload = true
	slog.Info("Parquet persisted successfully after delete")
	return nil
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
		// Parse and evaluate boolean expression
		tokens := tokenizeSearch(f.Search)
		if !evaluateExpression(tokens, 0, len(tokens), r, s.cfg.Columns, f.SearchCol) {
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

// Token types for search expression
type tokenType int

const (
	tokenWord tokenType = iota
	tokenOR
	tokenNOT
	tokenLParen
	tokenRParen
)

type token struct {
	typ   tokenType
	value string // for words
}

// tokenizeSearch splits the search query into tokens
func tokenizeSearch(query string) []token {
	var tokens []token
	words := strings.Fields(query)

	for _, word := range words {
		wordLower := strings.ToUpper(word)
		switch wordLower {
		case "OR":
			tokens = append(tokens, token{typ: tokenOR, value: word})
		case "(":
			tokens = append(tokens, token{typ: tokenLParen, value: word})
		case ")":
			tokens = append(tokens, token{typ: tokenRParen, value: word})
		default:
			// Check for NOT prefix (-)
			if strings.HasPrefix(word, "-") {
				tokens = append(tokens, token{typ: tokenNOT, value: word[1:]})
			} else {
				tokens = append(tokens, token{typ: tokenWord, value: word})
			}
		}
	}

	return tokens
}

// evaluateExpression evaluates a boolean expression from tokens[start:end]
// Returns true if the expression matches the row
func evaluateExpression(tokens []token, start, end int, row map[string]any, columns []config.ColumnDef, searchCol string) bool {
	if start >= end {
		return true
	}

	// Parse groups separated by OR
	// Each group is ANDed together, groups are ORed
	groupStart := start
	i := start

	for i < end {
		if tokens[i].typ == tokenOR {
			// Evaluate current group (AND logic)
			groupMatch := evaluateGroup(tokens, groupStart, i, row, columns, searchCol)
			if groupMatch {
				return true
			}
			// Start next group after OR
			groupStart = i + 1
		}
		i++
	}

	// Evaluate last group
	return evaluateGroup(tokens, groupStart, end, row, columns, searchCol)
}

// evaluateGroup evaluates terms within a group (AND logic)
func evaluateGroup(tokens []token, start, end int, row map[string]any, columns []config.ColumnDef, searchCol string) bool {
	i := start
	notActive := false

	for i < end {
		tok := tokens[i]

		if tok.typ == tokenNOT {
			notActive = true
			i++
			continue
		}

		if tok.typ == tokenLParen {
			// Find matching closing paren
			parenStart := i + 1
			parenEnd := findMatchingParen(tokens, i)
			groupMatch := evaluateExpression(tokens, parenStart, parenEnd, row, columns, searchCol)
			if notActive {
				groupMatch = !groupMatch
				notActive = false
			}
			if !groupMatch {
				return false
			}
			i = parenEnd + 1
			continue
		}

		if tok.typ == tokenWord {
			match := matchWord(row, columns, searchCol, tok.value)
			if notActive {
				match = !match
				notActive = false
			}
			if !match {
				return false
			}
		}

		i++
	}

	return true
}

// findMatchingParen finds the index of the closing parenthesis
func findMatchingParen(tokens []token, start int) int {
	count := 1
	i := start + 1
	for i < len(tokens) && count > 0 {
		if tokens[i].typ == tokenLParen {
			count++
		} else if tokens[i].typ == tokenRParen {
			count--
		}
		i++
	}
	return i - 1
}

// matchWord checks if a word matches the row (case-insensitive substring)
func matchWord(row map[string]any, columns []config.ColumnDef, searchCol, word string) bool {
	searchLower := strings.ToLower(word)

	if searchCol != "" {
		if val, ok := row[searchCol]; ok {
			return strings.Contains(strings.ToLower(fmt.Sprintf("%v", val)), searchLower)
		}
		return false
	}

	for _, col := range columns {
		if !col.Searchable {
			continue
		}
		if val, ok := row[col.Name]; ok {
			if strings.Contains(strings.ToLower(fmt.Sprintf("%v", val)), searchLower) {
				return true
			}
		}
	}
	return false
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
