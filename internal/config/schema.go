package config

import (
	"encoding/json"
	"fmt"
	"os"
)

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
