package gogetvers

import (
	"os"
	"path/filepath"
)

// Returns basename from path.
func Basename(path string) string {
	return filepath.Base(path)
}

// Dir returns all but the last element of path.
func Dir(path string) string {
	return filepath.Dir(path)
}

// Determines if path is a file; returns true if it is.
func IsFile(path string) bool {
	if len(path) == 0 {
		return false
	}
	finfo, err := os.Stat(path)
	return err == nil && !finfo.IsDir()

}

// Determines if path is a directory; returns true if it is.
func IsDir(path string) bool {
	if len(path) == 0 {
		return false
	}
	finfo, err := os.Stat(path)
	return err == nil && finfo.IsDir()
}

// Makes a directory and any necessary parents.
func Mkdir(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
