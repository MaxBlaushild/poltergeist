package util

import (
	"bytes"
	"fmt"
)

func DetectImageFormat(data []byte) (string, error) {
	// Common image format signatures
	signatures := map[string][]byte{
		"image/jpeg": {0xFF, 0xD8, 0xFF},
		"image/png":  {0x89, 0x50, 0x4E, 0x47},
		"image/gif":  {0x47, 0x49, 0x46, 0x38},
		"image/webp": {0x52, 0x49, 0x46, 0x46},
	}

	for mimeType, sig := range signatures {
		if len(data) >= len(sig) && bytes.Equal(data[:len(sig)], sig) {
			return mimeType, nil
		}
	}
	return "", fmt.Errorf("unknown image format: %s", data)
}

func GetImageExtension(mimeType string) (string, error) {
	extensions := map[string]string{
		"image/jpeg": "jpg",
		"image/png":  "png",
		"image/gif":  "gif",
		"image/webp": "webp",
	}

	if ext, ok := extensions[mimeType]; ok {
		return ext, nil
	}
	return "", fmt.Errorf("unknown mime type: %s", mimeType)
}
