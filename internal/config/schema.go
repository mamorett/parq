package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MultiConfig is the new multi-parquet config format
type MultiConfig struct {
	Parquets []ParquetEntry `json:"parquets"`
}

// ParquetEntry represents a single parquet file configuration
type ParquetEntry struct {
	Path        string     `json:"path"`
	Name        string     `json:"name,omitempty"`        // derived from basename if empty
	Columns     []ColumnDef `json:"columns,omitempty"`     // empty → auto-discover
	DefaultSort SortDef     `json:"default_sort,omitempty"`
	Pagination  Pagination  `json:"pagination,omitempty"`
	Thumbnail   Thumbnail   `json:"thumbnail,omitempty"`
}

// Resolve converts a ParquetEntry to a Config, auto-discovering columns if needed
func (e *ParquetEntry) Resolve() (*Config, error) {
	if len(e.Columns) == 0 {
		cfg, err := Discover(e.Path)
		if err != nil {
			return nil, err
		}
		// Apply entry-level overrides
		if e.Name != "" {
			// Name is metadata, not part of Config
		}
		if e.DefaultSort.Column != "" {
			cfg.DefaultSort = e.DefaultSort
		}
		if e.Pagination.DefaultPageSize != 0 {
			cfg.Pagination = e.Pagination
		}
		if e.Thumbnail.MaxSize != 0 {
			cfg.Thumbnail = e.Thumbnail
		}
		return cfg, nil
	}

	// Use explicit columns
	cfg := &Config{
		ParquetFile: e.Path,
		Columns:     e.Columns,
	}
	if e.DefaultSort.Column != "" {
		cfg.DefaultSort = e.DefaultSort
	}
	if e.Pagination.DefaultPageSize != 0 {
		cfg.Pagination = e.Pagination
	} else {
		cfg.Pagination = Pagination{
			DefaultPageSize: 10,
			PageSizeOptions: []int{5, 10, 20, 50, 100},
		}
	}
	if e.Thumbnail.MaxSize != 0 {
		cfg.Thumbnail = e.Thumbnail
	} else {
		cfg.Thumbnail = Thumbnail{MaxSize: 300, Format: "webp"}
	}
	return cfg, nil
}

// GetName returns the display name for this entry (derived from filename if not set)
func (e *ParquetEntry) GetName() string {
	if e.Name != "" {
		return e.Name
	}
	base := filepath.Base(e.Path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

type Config struct {
	ParquetFile string      `json:"parquet_file"`
	Columns     []ColumnDef `json:"columns"`
	DefaultSort SortDef     `json:"default_sort"`
	Pagination  Pagination  `json:"pagination"`
	Thumbnail   Thumbnail   `json:"thumbnail"`
}

type ColumnDef struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"` // string, int, blob, path
	Label           string   `json:"label"`
	Searchable      bool     `json:"searchable"`
	Editable        bool     `json:"editable"`
	Sortable        bool     `json:"sortable"`
	Copyable        bool     `json:"copyable"`
	Hidden          bool     `json:"hidden"`
	Format          string   `json:"format"` // e.g., "datetime"
	Remap           []Remap  `json:"remap"`
	ProbeDimensions bool     `json:"probe_dimensions"`
}

type SortDef struct {
	Column string `json:"column"`
	Order  string `json:"order"` // "asc" or "desc"
}

type Pagination struct {
	DefaultPageSize int   `json:"default_page_size"`
	PageSizeOptions []int `json:"page_size_options"`
}

type Thumbnail struct {
	Column  string `json:"column"`
	MaxSize int    `json:"max_size"`
	Format  string `json:"format"` // "webp" or "jpeg"
}

type Remap struct {
	Pattern string `json:"pattern"`
	Replace string `json:"replace"`
}

func Load(path, parquetOverride string, autoDiscover bool) (*Config, error) {
	var cfg Config
	
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			if autoDiscover && parquetOverride != "" {
				// We'll handle discovery outside or return a special error/signal
				return Discover(parquetOverride)
			}
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	if parquetOverride != "" {
		cfg.ParquetFile = parquetOverride
	}

	// Apply defaults
	if cfg.Pagination.DefaultPageSize == 0 {
		cfg.Pagination.DefaultPageSize = 10
	}
	if len(cfg.Pagination.PageSizeOptions) == 0 {
		cfg.Pagination.PageSizeOptions = []int{5, 10, 20, 50, 100}
	}
	if cfg.Thumbnail.MaxSize == 0 {
		cfg.Thumbnail.MaxSize = 300
	}
	if cfg.Thumbnail.Format == "" {
		cfg.Thumbnail.Format = "webp"
	}

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(c)
}

// LoadMulti loads a multi-parquet config file (parqs.json)
func LoadMulti(path string) (*MultiConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, err
	}
	defer f.Close()

	var mc MultiConfig
	if err := json.NewDecoder(f).Decode(&mc); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Resolve each entry (auto-discover if columns are empty)
	for i := range mc.Parquets {
		entry := &mc.Parquets[i]
		if len(entry.Columns) == 0 {
			// Will be resolved on demand by MultiStore
			// Just set the name here
			if entry.Name == "" {
				entry.Name = entry.GetName()
			}
		}
	}

	return &mc, nil
}

// LoadLegacy loads a single-file schema.json and wraps it in a MultiConfig
// This is for backward compatibility
func LoadLegacy(path, parquetOverride string, autoDiscover bool) (*MultiConfig, error) {
	var cfg Config

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			if autoDiscover && parquetOverride != "" {
				discovered, err := Discover(parquetOverride)
				if err != nil {
					return nil, err
				}
				return &MultiConfig{
					Parquets: []ParquetEntry{
						{
							Path:   parquetOverride,
							Name:   filepath.Base(parquetOverride),
							Columns: discovered.Columns,
						},
					},
				}, nil
			}
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	if parquetOverride != "" {
		cfg.ParquetFile = parquetOverride
	}

	// Apply defaults
	if cfg.Pagination.DefaultPageSize == 0 {
		cfg.Pagination.DefaultPageSize = 10
	}
	if len(cfg.Pagination.PageSizeOptions) == 0 {
		cfg.Pagination.PageSizeOptions = []int{5, 10, 20, 50, 100}
	}
	if cfg.Thumbnail.MaxSize == 0 {
		cfg.Thumbnail.MaxSize = 300
	}
	if cfg.Thumbnail.Format == "" {
		cfg.Thumbnail.Format = "webp"
	}

	// Wrap in MultiConfig
	base := filepath.Base(cfg.ParquetFile)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return &MultiConfig{
		Parquets: []ParquetEntry{
			{
				Path:   cfg.ParquetFile,
				Name:   name,
				Columns: cfg.Columns,
			},
		},
	}, nil
}
