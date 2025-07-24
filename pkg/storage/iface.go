package storage

import "github.com/callumj/clicky/pkg/config"

type Storage interface {
	SaveSnapshot(camera *config.CameraConfig, data []byte) error
}
