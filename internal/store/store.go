package store

import (
	"github.com/trithemius/parq/internal/image"
	"github.com/trithemius/parq/internal/stats"
)

type Row struct {
	Index     int              `json:"index"`
	Columns   map[string]any   `json:"columns"`
	ImageMeta *image.ImageMeta `json:"image_meta,omitempty"`
}

type Filter struct {
	Search    string            `json:"search"`
	SearchCol string            `json:"search_col"`
	Exact     map[string]string `json:"exact"`
	Subdirs   []string          `json:"subdirs"`
}

type Sort struct {
	Column string `json:"column"`
	Order  string `json:"order"` // "asc", "desc"
}

type Pagination struct {
	Page int `json:"page"`
	Size int `json:"size"`
}



type RowStore interface {
	Query(f Filter, s Sort, p Pagination) ([]Row, int, error)
	Get(idx int) (Row, error)
	Update(idx int, cols map[string]any) error
	Delete(idx int) error
	Stats() (stats.Stats, error)
	Subdirs(col string) ([]string, error)
	Close() error
}
