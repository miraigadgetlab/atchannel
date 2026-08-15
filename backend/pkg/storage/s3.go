package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 implements Storage against any S3-compatible backend
// (MinIO self-hosted, Cloudflare R2, Backblaze B2, AWS S3).
type S3 struct {
	client  *minio.Client
	bucket  string
	public  string // optional public base URL; empty means bucket URL
}

type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Region          string
	UseSSL          bool
	PublicBaseURL   string
}

func NewS3(cfg S3Config) (*S3, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentialsFor(cfg),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: minio client: %w", err)
	}
	return &S3{
		client: client,
		bucket: cfg.Bucket,
		public: strings.TrimSuffix(cfg.PublicBaseURL, "/"),
	}, nil
}

func credentialsFor(cfg S3Config) *credentials.Credentials {
	return credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, "")
}

func (s *S3) Put(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("storage: put: %w", err)
	}
	return nil
}

func (s *S3) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("storage: get: %w", err)
	}
	// Detect missing objects lazily.
	var probe [1]byte
	if _, err := obj.Read(probe[:]); err != nil {
		obj.Close()
		if err == io.EOF {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &prefixedReader{reader: io.MultiReader(bytes.NewReader(probe[:]), obj), closer: obj}, nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("storage: delete: %w", err)
	}
	return nil
}

func (s *S3) URL(ctx context.Context, key string) (string, error) {
	if s.public != "" {
		return s.public + "/" + key, nil
	}
	// Fall back to a signed URL; callers wanting cacheable public URLs
	// should configure PublicBaseURL.
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, 24*60*60, nil)
	if err != nil {
		return "", fmt.Errorf("storage: presign: %w", err)
	}
	return u.String(), nil
}

func (s *S3) Stat(ctx context.Context, key string) (int64, error) {
	info, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		var respErr minio.ErrorResponse
		if errors.As(err, &respErr) && respErr.Code == "NoSuchKey" {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("storage: stat: %w", err)
	}
	return info.Size, nil
}

type prefixedReader struct {
	reader io.Reader
	closer io.Closer
}

func (p *prefixedReader) Read(b []byte) (int, error) { return p.reader.Read(b) }
func (p *prefixedReader) Close() error               { return p.closer.Close() }
