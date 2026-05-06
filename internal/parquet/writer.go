package parquet

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/parquet-go/parquet-go"
)

func WriteAll(path string, data []map[string]any, schema *parquet.Schema) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "parq-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	writer := parquet.NewGenericWriter[map[string]any](tmp, schema)

	_, writeErr := writer.Write(data)
	closeErr := writer.Close()
	tmp.Close()

	if writeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("write parquet rows: %w", writeErr)
	}
	if closeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("flush parquet writer: %w", closeErr)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp to parquet: %w", err)
	}

	return nil
}
