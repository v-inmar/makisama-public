package image_utils

import (
	"os"
)

func GetExtFromContentType(contentType string) string {

	ext := ""
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		ext = ""
	}
	return ext
}

func DeleteImageFile(filepath string) error {
	if err := os.Remove(filepath); err != nil {
		return err
	}
	return nil
}
