package config

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/parquet-go/parquet-go"
)

var (
	pathSeparators = regexp.MustCompile(`[/\\|]`)
	imageExtensions = map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
		".bmp": true, ".tiff": true, ".mp4": true, ".mov": true, ".avi": true,
		".wav": true, ".mp3": true, ".txt": true, ".json": true, ".csv": true,
		".parquet": true, ".pdf": true,
	}
	// Common date formats
	dateFormats = []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02-Jan-2006",
	}
)

func Discover(parquetPath string) (*Config, error) {
	info, err := os.Stat(parquetPath)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(parquetPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pf, err := parquet.OpenFile(f, info.Size())
	if err != nil {
		return nil, err
	}
	schema := pf.Schema()

	cfg := &Config{
		ParquetFile: parquetPath,
		Pagination: Pagination{
			DefaultPageSize: 10,
			PageSizeOptions: []int{5, 10, 20, 50, 100},
		},
		Thumbnail: Thumbnail{
			MaxSize: 300,
			Format:  "webp",
		},
	}

	// Sample rows for heuristics
	sampleSize := 25
	reader := parquet.NewGenericReader[map[string]any](f, schema)
	defer reader.Close()

	samples := make([]map[string]any, 0, sampleSize)
	for i := 0; i < sampleSize; i++ {
		row := make(map[string]any)
		n, err := reader.Read([]map[string]any{row})
		if n > 0 {
			samples = append(samples, row)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	for _, field := range schema.Fields() {
		col := ColumnDef{
			Name:       field.Name(),
			Label:      field.Name(),
			Searchable: true,
			Sortable:   true,
			Copyable:   true,
		}

		detectType(&col, field, samples)
		cfg.Columns = append(cfg.Columns, col)
	}

	// Set default sort and thumbnail
	for _, col := range cfg.Columns {
		if col.Format == "datetime" {
			cfg.DefaultSort = SortDef{Column: col.Name, Order: "desc"}
			break
		}
	}
	if cfg.DefaultSort.Column == "" && len(cfg.Columns) > 0 {
		cfg.DefaultSort = SortDef{Column: cfg.Columns[0].Name, Order: "asc"}
	}

	for _, col := range cfg.Columns {
		if col.Type == "path" {
			cfg.Thumbnail.Column = col.Name
			break
		}
	}

	return cfg, nil
}

func detectType(col *ColumnDef, field parquet.Field, samples []map[string]any) {
	t := field.Type()
	
	// Basic types from Parquet schema
	switch {
	case t.Kind() == parquet.Int32 || t.Kind() == parquet.Int64:
		col.Type = "int"
	case t.Kind() == parquet.Float || t.Kind() == parquet.Double:
		col.Type = "string" // Renders as string for now
	case t.Kind() == parquet.Boolean:
		col.Type = "string"
	default:
		col.Type = "string"
	}

	// Heuristics for BYTE_ARRAY (Strings/Blobs/Paths)
	if t.Kind() == parquet.ByteArray || t.Kind() == parquet.FixedLenByteArray {
		// Sample based detection
		pathCount := 0
		imageExtCount := 0
		dateCount := 0
		nonNullCount := 0

		for _, sample := range samples {
			val, ok := sample[col.Name]
			if !ok || val == nil {
				continue
			}
			str, isStr := val.(string)
			if !isStr {
				// Might be []byte
				if b, isBytes := val.([]byte); isBytes {
					str = string(b)
				} else {
					continue
				}
			}
			if str == "" {
				continue
			}
			nonNullCount++

			// Path detection
			if pathSeparators.MatchString(str) {
				pathCount++
			}
			ext := strings.ToLower(filepath.Ext(str))
			if imageExtensions[ext] {
				imageExtCount++
			}

			// Date detection
			for _, fmtStr := range dateFormats {
				if _, err := time.Parse(fmtStr, str); err == nil {
					dateCount++
					break
				}
			}
		}

		if nonNullCount > 0 {
			if float64(pathCount)/float64(nonNullCount) >= 0.8 && float64(imageExtCount)/float64(nonNullCount) >= 0.5 {
				col.Type = "path"
				col.ProbeDimensions = true
			} else if float64(dateCount)/float64(nonNullCount) >= 0.8 {
				col.Type = "string"
				col.Format = "datetime"
			}
		}

		// Check if it's binary blob
		if col.Type == "string" && col.Format == "" {
			isBinary := false
			for _, sample := range samples {
				if val, ok := sample[col.Name]; ok {
					if b, isBytes := val.([]byte); isBytes {
						// Check if it contains null bytes or non-printable chars
						for _, bval := range b {
							if bval < 32 && bval != '\n' && bval != '\r' && bval != '\t' {
								isBinary = true
								break
							}
						}
					}
				}
				if isBinary {
					break
				}
			}
			if isBinary {
				col.Type = "blob"
				col.Sortable = false
			}
		}
	}

	if col.Type == "string" && col.Format != "datetime" {
		col.Editable = true
	}

	// Heuristic for numeric dates (INT32/INT64)
	if col.Type == "int" {
		dateCount := 0
		nonNullCount := 0
		for _, sample := range samples {
			val, ok := sample[col.Name]
			if !ok || val == nil {
				continue
			}
			var num int64
			switch v := val.(type) {
			case int64:
				num = v
			case int32:
				num = int64(v)
			case int:
				num = int64(v)
			default:
				continue
			}

			nonNullCount++
			// Heuristic: seconds, millis, micros, or nanos (2000-3000)
			// Year 2000 (seconds): 946,684,800
			// Year 3000 (nanos): 32,503,680,000,000,000,000 (approx)
			if num >= 946684800 && num <= 4000000000000000000 {
				dateCount++
			}
		}
		if nonNullCount > 0 && float64(dateCount)/float64(nonNullCount) >= 0.8 {
			col.Format = "datetime"
		}
	}
}
