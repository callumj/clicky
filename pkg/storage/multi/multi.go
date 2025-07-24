package multi

import (
	"fmt"

	"github.com/callumj/clicky/pkg/config"
	"github.com/callumj/clicky/pkg/storage"
	"github.com/rs/zerolog/log"
)

type MultiStorage struct {
	backends map[string]storage.Storage
}

func NewMultiStorage() *MultiStorage {
	return &MultiStorage{
		backends: make(map[string]storage.Storage),
	}
}

func (m *MultiStorage) Register(name string, strg storage.Storage) {
	log.Info().Str("storage", name).Msg("Registering storage backend")
	m.backends[name] = strg
}

func (m *MultiStorage) SaveSnapshot(camera *config.CameraConfig, data []byte) error {
	for name, s := range m.backends {
		if err := s.SaveSnapshot(camera, data); err != nil {
			return fmt.Errorf("failed to save snapshot for camera %s using storage %s: %w", camera.Name, name, err)
		}
	}

	return nil
}
