package files

import (
	"github.com/samedi/caldav-go/lib"
	"path/filepath"
	"strings"
)

const (
	Separator = string(filepath.Separator)
)

// AbsPath converts the path into absolute path based on the current working directory.
func AbsPath(path string) string {
	path = strings.Trim(path, "/")
	absPath, _ := filepath.Abs(path)

	return absPath
}

// DirPath returns all but the last element of path, typically the path's directory.
func DirPath(path string) string {
	return filepath.Dir(path)
}

// JoinPaths joins two or more paths into a single path.
func JoinPaths(paths ...string) string {
	return filepath.Join(paths...)
}

// ToSlashPath slashify the path, using '/' as separator.
func ToSlashPath(path string) string {
	return lib.ToSlashPath(path)
}
