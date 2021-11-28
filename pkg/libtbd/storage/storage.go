package storage

import (
	"io"
	"os"
	"path/filepath"
)

type Storage interface {
	Upload(reader io.ReadCloser) (string, error)
	Download(checksum string) (io.ReadCloser, error)
	Path(checksum string) (string, error)
}

func EnsureParentDir(subdir string) (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	parentDir := filepath.FromSlash(dir + "/tbd/storage/" + subdir)
	err = os.MkdirAll(parentDir, 0755)
	if err != nil {
		return "", err
	}
	return parentDir, nil
}
