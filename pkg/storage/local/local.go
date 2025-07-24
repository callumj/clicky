package local

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/callumj/clicky/pkg/config"
	"github.com/callumj/clicky/pkg/storage"
)

type LocalStorage struct {
	Path string
}

func NewLocalStorage(config *config.LocalStorageConfig) *LocalStorage {
	return &LocalStorage{
		Path: config.Path,
	}
}

func (s *LocalStorage) SaveSnapshot(camera *config.CameraConfig, data []byte) error {
	path := storage.PathForSnapshot(camera)
	fullPath := filepath.Join(s.Path, path)

	dir := filepath.Dir(fullPath)
	// check the directory exists, if not create it
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	mimeType := http.DetectContentType(data)
	fileExt, err := mime.ExtensionsByType(mimeType)
	if len(fileExt) > 0 && err == nil {
		fullPath += fileExt[len(fileExt)-1] // use the last extension if multiple are returned
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	_, err = file.Write(data)
	return err
}
