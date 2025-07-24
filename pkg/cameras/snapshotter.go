package cameras

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/callumj/clicky/pkg/config"
	"github.com/callumj/clicky/pkg/storage"
)

var (
	errEmptySnapshotURL = errors.New("empty snapshot URL")
)

type Snapshotter struct {
	client  *http.Client
	config  *config.RootConfig
	storage storage.Storage
}

func NewSnapshotterWithClient(client *http.Client, config *config.RootConfig, storage storage.Storage) *Snapshotter {
	if client == nil {
		client = http.DefaultClient
	}
	return &Snapshotter{
		client:  client,
		config:  config,
		storage: storage,
	}
}

func NewSnapshotter(config *config.RootConfig, storage storage.Storage) *Snapshotter {
	return &Snapshotter{
		client:  http.DefaultClient,
		config:  config,
		storage: storage,
	}
}

func (s *Snapshotter) SaveSnapshots() error {
	for _, cam := range s.config.Cameras {
		data, err := s.getSnapshot(cam)
		if err != nil {
			return fmt.Errorf("failed to get snapshot for camera %s: %w", cam.Name, err)
		}

		if err := s.storage.SaveSnapshot(cam, data); err != nil {
			return fmt.Errorf("failed to save snapshot for camera %s: %w", cam.Name, err)
		}
	}
	return nil
}

func (s *Snapshotter) getSnapshot(camera *config.CameraConfig) ([]byte, error) {
	if camera.SnapshotURL == "" {
		return nil, errEmptySnapshotURL
	}

	resp, err := s.client.Get(camera.SnapshotURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get snapshot: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
