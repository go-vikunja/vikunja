package lib

import (
	"path/filepath"
)

func ToSlashPath(path string) string {
	cleanPath := filepath.Clean(path)
	return filepath.ToSlash(cleanPath)
}
