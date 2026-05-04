package stats

import (
	"os"

	"github.com/trithemius/parq/internal/config"
)

type Stats struct {
	TotalRows     int            `json:"total_rows"`
	ImagesFound   int            `json:"images_found"`
	ImagesMissing int            `json:"images_missing"`
	DateRange     map[string]any `json:"date_range"`
	FileSize      int64          `json:"file_size_bytes"`
}

func Compute(data []map[string]any, cfg *config.Config) Stats {
	stats := Stats{
		TotalRows: len(data),
		DateRange: make(map[string]any),
	}

	pathCol := ""
	if cfg != nil {
		for _, c := range cfg.Columns {
			if c.Type == "path" {
				pathCol = c.Name
				break
			}
		}
	}

	found := 0
	missing := 0
	for _, r := range data {
		if pathCol != "" {
			if p, ok := r[pathCol].(string); ok {
				if _, err := os.Stat(p); err == nil {
					found++
				} else {
					missing++
				}
			}
		}
	}

	stats.ImagesFound = found
	stats.ImagesMissing = missing

	return stats
}
