package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Local implements Storage against a directory on local disk.
type Local struct {
	baseDir string // absolute or relative on-disk root
	baseURL string // public URL prefix, e.g. /files
}

func NewLocal(baseDir, baseURL string) (*Local, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &Local{baseDir: baseDir, baseURL: strings.TrimSuffix(baseURL, "/")}, nil
}

func (l *Local) path(key string) (string, error) {
	clean := filepath.Clean(key)
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("storage: invalid key %q", key)
	}
	return filepath.Join(l.baseDir, clean), nil
}

func (l *Local) Put(ctx context.Context, key string, data []byte, contentType string) error {
	p, err := l.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("storage: mkdir: %w", err)
	}
	return os.WriteFile(p, data, 0o644)
}

func (l *Local) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	p, err := l.path(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("storage: open: %w", err)
	}
	return f, nil
}

func (l *Local) Delete(ctx context.Context, key string) error {
	p, err := l.path(key)
	if err != nil {
		return err
	}
	err = os.Remove(p)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (l *Local) URL(ctx context.Context, key string) (string, error) {
	return l.baseURL + "/" + key, nil
}

func (l *Local) Stat(ctx context.Context, key string) (int64, error) {
	p, err := l.path(key)
	if err != nil {
		return 0, err
	}
	fi, err := os.Stat(p)
	if os.IsNotExist(err) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}
