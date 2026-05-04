package parquet

import (
	"os"

	"github.com/parquet-go/parquet-go"
)

func WriteAll(path string, data []map[string]any, schema *parquet.Schema) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := parquet.NewGenericWriter[map[string]any](f, schema)
	defer writer.Close()

	_, err = writer.Write(data)
	return err
}
