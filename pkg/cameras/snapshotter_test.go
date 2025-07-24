package cameras

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/callumj/clicky/pkg/config"
)

type dummyStorage struct{}

func (d *dummyStorage) SaveSnapshot(cam *config.CameraConfig, data []byte) error { return nil }

func TestGetSnapshot_EmptyURL(t *testing.T) {
	s := NewSnapshotter(nil, &dummyStorage{})
	cam := &config.CameraConfig{Name: "TestCam", SnapshotURL: ""}
	_, err := s.getSnapshot(cam)
	if !errors.Is(err, errEmptySnapshotURL) {
		t.Errorf("expected errEmptySnapshotURL, got %v", err)
	}
}

func TestGetSnapshot_HTTPErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := NewSnapshotter(nil, &dummyStorage{})
	cam := &config.CameraConfig{Name: "TestCam", SnapshotURL: ts.URL}
	_, err := s.getSnapshot(cam)
	if err == nil || err.Error() == "" || !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error containing 500, got %v", err)
	}
}

func TestGetSnapshot_Success(t *testing.T) {
	expected := []byte("snapshotdata")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(expected)
	}))
	defer ts.Close()

	s := NewSnapshotter(nil, &dummyStorage{})
	cam := &config.CameraConfig{Name: "TestCam", SnapshotURL: ts.URL}
	data, err := s.getSnapshot(cam)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(data) != string(expected) {
		t.Errorf("expected %q, got %q", expected, data)
	}
}

func TestGetSnapshot_ClientError(t *testing.T) {
	s := NewSnapshotter(nil, &dummyStorage{})
	cam := &config.CameraConfig{Name: "TestCam", SnapshotURL: "http://invalid.invalid"}
	_, err := s.getSnapshot(cam)
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}
