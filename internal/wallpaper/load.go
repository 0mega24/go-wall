package wallpaper

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

// detectFormat reads up to 12 bytes from r, identifies the image format by
// magic bytes, and returns the format name together with a repaired reader
// that still contains those bytes.
func detectFormat(r io.Reader) (string, io.Reader, error) {
	header := make([]byte, 12)
	n, err := io.ReadFull(r, header)
	header = header[:n]
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", nil, err
	}
	full := io.MultiReader(bytes.NewReader(header), r)
	switch {
	case len(header) >= 4 && header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47:
		return "png", full, nil
	case len(header) >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF:
		return "jpeg", full, nil
	case len(header) >= 12 && header[0] == 'R' && header[1] == 'I' && header[2] == 'F' && header[3] == 'F' &&
		header[8] == 'W' && header[9] == 'E' && header[10] == 'B' && header[11] == 'P':
		return "webp", full, nil
	case len(header) >= 4 && ((header[0] == 0x49 && header[1] == 0x49 && header[2] == 0x2A && header[3] == 0x00) ||
		(header[0] == 0x4D && header[1] == 0x4D && header[2] == 0x00 && header[3] == 0x2A)):
		return "tiff", full, nil
	case len(header) >= 2 && header[0] == 0x42 && header[1] == 0x4D:
		return "bmp", full, nil
	default:
		return "", full, fmt.Errorf("gowall: unrecognized image format")
	}
}

// LoadImage decodes an image file (JPEG, PNG, WebP, TIFF, BMP) and returns it.
func LoadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening image: %w", err)
	}
	defer func() { _ = f.Close() }()

	format, r, err := detectFormat(f)
	if err != nil {
		return nil, fmt.Errorf("detecting format: %w", err)
	}

	switch format {
	case "png":
		return png.Decode(r)
	case "jpeg":
		return jpeg.Decode(r)
	case "webp":
		return webp.Decode(r)
	case "tiff":
		return tiff.Decode(r)
	case "bmp":
		return bmp.Decode(r)
	default:
		return nil, fmt.Errorf("gowall: unsupported format %q", format)
	}
}
