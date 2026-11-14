package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog"
	"ielts-practice/app/modules/storage/ports"
	"ielts-practice/configs"
)

var ErrNotConfigured = errors.New("object storage is not configured (S3_BUCKET_NAME is empty)")

var _ ports.IObjectStorage = &S3Storage{}

type S3Storage struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

func NewS3Storage(cfg *configs.Config, logger zerolog.Logger) (ports.IObjectStorage, error) {
	sc := cfg.Storage
	if sc.Bucket == "" {
		logger.Warn().Msg("S3_BUCKET_NAME is empty: media storage disabled")
		return &S3Storage{}, nil
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(sc.Region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if sc.Endpoint != "" {
			o.BaseEndpoint = aws.String(sc.Endpoint)
		}
		o.UsePathStyle = sc.UsePathStyle
	})
	return &S3Storage{client: client, presign: s3.NewPresignClient(client), bucket: sc.Bucket}, nil
}

func (s *S3Storage) Configured() bool { return s.client != nil }

func (s *S3Storage) Ping(ctx context.Context) error {
	if !s.Configured() {
		return ErrNotConfigured
	}
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)})
	return err
}

func (s *S3Storage) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if !s.Configured() {
		return ErrNotConfigured
	}
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
		// Media keys are content-addressed, so they never change once written.
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	return err
}

func (s *S3Storage) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if !s.Configured() {
		return "", ErrNotConfigured
	}
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	if !s.Configured() {
		return ErrNotConfigured
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	return err
}
