package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage struct {
	client *s3.Client
	bucket string
}

func newStorage(ctx context.Context) (*Storage, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return &Storage{
		client: s3.NewFromConfig(cfg),
		bucket: os.Getenv("S3_BUCKET"),
	}, nil
}

func (s *Storage) uploadPoster(ctx context.Context, fileID, localPath string) (string, error) {
	file, err := os.Open(localPath) //#nosec G304 -- localPath is built from internal fileID, not user input
	if err != nil {
		return "", fmt.Errorf("open poster: %w", err)
	}
	defer file.Close()

	key := fmt.Sprintf("posters/%s.png", fileID)

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String("image/png"),
	})
	if err != nil {
		return "", fmt.Errorf("upload to s3: %w", err)
	}

	presignClient := s3.NewPresignClient(s.client)
	presigned, err := presignClient.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(24*time.Hour),
	)
	if err != nil {
		return "", fmt.Errorf("presign url: %w", err)
	}

	return presigned.URL, nil
}
