package config

import (
	"os"
	"testing"
)

func TestSnapshotConfig_CronSchedule(t *testing.T) {
	expected := "0 0 * * *"
	cfg := SnapshotConfig{CronSchedule: expected}
	if cfg.CronSchedule != expected {
		t.Errorf("expected CronSchedule %q, got %q", expected, cfg.CronSchedule)
	}
}

func TestLoadConfig_Success(t *testing.T) {
	content := `
cameras:
  - name: "Cam1"
    snapshot_url: "http://example.com/cam1.jpg"
snapshot:
  cron_schedule: "0 5 * * *"
`
	tmpfile, err := os.CreateTemp("", "testconfig*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	go func() { _ = tmpfile.Close() }()

	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if len(cfg.Cameras) != 1 {
		t.Errorf("expected 1 camera, got %d", len(cfg.Cameras))
	}
	if cfg.Cameras[0].Name != "Cam1" {
		t.Errorf("expected camera name 'Cam1', got %q", cfg.Cameras[0].Name)
	}
	if cfg.Snapshot == nil {
		t.Fatal("expected Snapshot to be non-nil")
	}
	if cfg.Snapshot.CronSchedule != "0 5 * * *" {
		t.Errorf("expected cron_schedule '0 5 * * *', got %q", cfg.Snapshot.CronSchedule)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}
