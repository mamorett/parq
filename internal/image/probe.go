package image

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	_ "golang.org/x/image/webp"
)

type ImageMeta struct {
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Aspect     string  `json:"aspect"`
	Megapixels float64 `json:"megapixels"`
	FileSizeKB float64 `json:"file_size_kb"`
}

func Probe(path string) (*ImageMeta, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	config, _, err := image.DecodeConfig(f)
	if err != nil {
		return nil, err
	}

	meta := &ImageMeta{
		Width:      config.Width,
		Height:     config.Height,
		FileSizeKB: float64(info.Size()) / 1024,
		Megapixels: float64(config.Width*config.Height) / 1000000,
	}

	if config.Height > 0 {
		meta.Aspect = fmt.Sprintf("%d:%d", config.Width, config.Height)
	}

	return meta, nil
}
