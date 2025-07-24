package storage

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/callumj/clicky/pkg/config"
)

func PathForSnapshot(camera *config.CameraConfig) string {
	curTime := time.Now().Local()

	return filepath.Join("snapshots", pathForCamera(camera), strconv.Itoa(curTime.Year()), strconv.Itoa(int(curTime.Month())), strconv.Itoa(curTime.Day()), strconv.Itoa(int(curTime.Unix())))
}

func pathForCamera(camera *config.CameraConfig) string {
	return strings.ToLower(camera.Name)
}
