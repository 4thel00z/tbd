package storage

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	parentDir string
	registry  map[string]string
}

func NewLocalStorage() (LocalStorage, error) {
	parentDir, err := EnsureParentDir("local")
	if err != nil {
		return LocalStorage{}, err
	}
	return LocalStorage{
		parentDir: parentDir,
		registry:  map[string]string{},
	}, nil
}

func (l LocalStorage) Upload(reader io.ReadCloser) (string, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	uuidStr := u.String()
	clean := filepath.Clean(filepath.FromSlash(l.parentDir + "/" + uuidStr))
	l.registry[uuidStr] = clean
	f, err := os.Create(clean)

	if err != nil {
		return "", err
	}

	_, err = io.Copy(f, reader)
	if err != nil {
		return "", err
	}
	err = os.Chmod(clean, 0755)
	if err != nil {
		return "", err
	}
	return uuidStr, nil
}

func (l LocalStorage) Download(checksum string) (io.ReadCloser, error) {
	path, err := l.Path(checksum)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (l LocalStorage) Path(checksum string) (string, error) {
	path, ok := l.registry[checksum]
	if !ok {
		return "", fmt.Errorf("%s not found in download registry", checksum)
	}
	return path, nil
}
