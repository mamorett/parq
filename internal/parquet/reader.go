package parquet

import (
	"io"
	"os"

	"github.com/parquet-go/parquet-go"
)

func ReadAll(path string) ([]map[string]any, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pf, err := parquet.OpenFile(f, info.Size())
	if err != nil {
		return nil, err
	}

	reader := parquet.NewGenericReader[map[string]any](f, pf.Schema())
	defer reader.Close()

	var allRows []map[string]any
	for {
		row := make(map[string]any)
		rows, err := reader.Read([]map[string]any{row})
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if rows == 0 {
			break
		}
		allRows = append(allRows, row)
	}

	return allRows, nil
}
