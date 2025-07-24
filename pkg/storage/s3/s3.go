package s3

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	clickyConfig "github.com/callumj/clicky/pkg/config"
	"github.com/callumj/clicky/pkg/storage"
)

type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3Storage(cfg *clickyConfig.S3StorageConfig) (*S3Storage, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}
	return &S3Storage{
		client: s3.NewFromConfig(awsCfg),
		bucket: cfg.Bucket,
	}, nil
}

func (s *S3Storage) SaveSnapshot(camera *clickyConfig.CameraConfig, data []byte) error {
	fullPath := storage.PathForSnapshot(camera)

	mimeType := http.DetectContentType(data)
	fileExt, err := mime.ExtensionsByType(mimeType)
	if len(fileExt) > 0 && err == nil {
		fullPath += fileExt[len(fileExt)-1] // use the last extension if multiple are returned
	}

	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(fullPath),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(mimeType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}
	return nil
}
