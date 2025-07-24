package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type RootConfig struct {
	Cameras  []*CameraConfig `yaml:"cameras"`
	Snapshot *SnapshotConfig `yaml:"snapshot,omitempty"`
	Storage  *StorageConfig  `yaml:"storage,omitempty"`
}
type StorageConfig struct {
	Local *LocalStorageConfig `yaml:"local,omitempty"`
	S3    *S3StorageConfig    `yaml:"s3,omitempty"`
}

type LocalStorageConfig struct {
	Path string `yaml:"path,omitempty"`
}

type S3StorageConfig struct {
	Bucket          string `yaml:"bucket,omitempty"`
	Region          string `yaml:"region,omitempty"`
	AccessKeyID     string `yaml:"access_key_id,omitempty"`
	SecretAccessKey string `yaml:"secret_access_key,omitempty"`
}

type CameraConfig struct {
	Name        string `yaml:"name"`
	SnapshotURL string `yaml:"snapshot_url,omitempty"`
}

type SnapshotConfig struct {
	CronSchedule string `yaml:"cron_schedule"`
}

func LoadConfig(filepath string) (*RootConfig, error) {
	var config RootConfig

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
