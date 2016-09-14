package gogetvers

import (
	"os"
)

// IsFile determines if path is a file; returns true if it is.
func IsFile(path string) bool {
	if len(path) == 0 {
		return false
	}
	finfo, err := os.Stat(path)
	return err == nil && !finfo.IsDir()

}

// IsDir determines if path is a directory; returns true if it is.
func IsDir(path string) bool {
	if len(path) == 0 {
		return false
	}
	finfo, err := os.Stat(path)
	return err == nil && finfo.IsDir()
}

// Mkdir makes a directory and any necessary parents.
func Mkdir(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
