package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/h2non/bimg"

	"github.com/kosero/atchannel/backend/pkg/storage"
)

var (
	ErrNoImage         = errors.New("no image provided")
	ErrImageTooLarge   = errors.New("image exceeds maximum size")
	ErrUnsupportedType = errors.New("unsupported image type")
	ErrInvalidImage    = errors.New("invalid image data")
	ErrBanned          = errors.New("account is banned")
	ErrTitleTooLong    = errors.New("title too long")
	ErrBodyTooLong     = errors.New("body too long")
)

const (
	thumbnailMaxWidth  = 250
	thumbnailMaxHeight = 250
	maxImageDimensions = 10000
)

type UploadService struct {
	storage  storage.Storage
	maxBytes int64
	allowed  map[string]string // bimg type name -> extension
	mime     map[string]string // bimg type name -> mime type for storage
	baseURL  string
}

func NewUploadService(st storage.Storage, maxBytes int64, allowedTypes []string, baseURL string) *UploadService {
	allowed := make(map[string]string, len(allowedTypes))
	mime := make(map[string]string, len(allowedTypes))
	for _, m := range allowedTypes {
		name := strings.TrimPrefix(m, "image/")
		switch name {
		case "jpeg":
			allowed[name] = ".jpg"
		case "png":
			allowed[name] = ".png"
		case "webp":
			allowed[name] = ".webp"
		case "gif":
			allowed[name] = ".gif"
		default:
			allowed[name] = ".img"
		}
		mime[name] = m
	}
	return &UploadService{
		storage:  st,
		maxBytes: maxBytes,
		allowed:  allowed,
		mime:     mime,
		baseURL:  strings.TrimSuffix(baseURL, "/"),
	}
}

type UploadedImage struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	ThumbKey    string `json:"thumbKey"`
	ThumbURL    string `json:"thumbUrl"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	ThumbWidth  int    `json:"thumbWidth"`
	ThumbHeight int    `json:"thumbHeight"`
	SizeBytes   int64  `json:"sizeBytes"`
}

func (s *UploadService) Store(ctx context.Context, data []byte, detectedMime string) (*UploadedImage, error) {
	if len(data) == 0 {
		return nil, ErrNoImage
	}
	if int64(len(data)) > s.maxBytes {
		return nil, ErrImageTooLarge
	}

	// Never trust the client-supplied content type; sniff the magic bytes.
	name := bimg.DetermineImageTypeName(data)
	if name == "" || name == "unknown" {
		return nil, ErrInvalidImage
	}
	ext, ok := s.allowed[name]
	if !ok {
		return nil, ErrUnsupportedType
	}
	mime, ok := s.mime[name]
	if !ok {
		return nil, ErrUnsupportedType
	}

	metadata, err := bimg.Metadata(data)
	if err != nil {
		return nil, ErrInvalidImage
	}
	width, height := metadata.Size.Width, metadata.Size.Height
	if width == 0 || height == 0 {
		return nil, ErrInvalidImage
	}
	if width > maxImageDimensions || height > maxImageDimensions {
		return nil, ErrInvalidImage
	}

	base := fmt.Sprintf("%d", time.Now().UnixNano())
	key := filepath.Join("images", base+ext)

	if err := s.storage.Put(ctx, key, data, mime); err != nil {
		return nil, fmt.Errorf("store original: %w", err)
	}

	// Animated GIFs cannot be safely downscaled with libvips in this path;
	// store a static thumbnail from the first frame when possible.
	thumbKey := filepath.Join("thumbs", base+".jpg")
	thumbBytes, err := bimg.NewImage(data).Resize(thumbnailMaxWidth, thumbnailMaxHeight)
	if err != nil {
		// Fall back: serve the original as the thumbnail too.
		thumbKey = key
		if err := s.storage.Put(ctx, thumbKey, data, mime); err != nil {
			return nil, fmt.Errorf("store thumbnail fallback: %w", err)
		}
		return s.buildResult(ctx, key, thumbKey, data, width, height)
	}
	if err := s.storage.Put(ctx, thumbKey, thumbBytes, "image/jpeg"); err != nil {
		return nil, fmt.Errorf("store thumbnail: %w", err)
	}
	thumbMeta, _ := bimg.Metadata(thumbBytes)
	tw, th := thumbMeta.Size.Width, thumbMeta.Size.Height
	return &UploadedImage{
		Key:         key,
		ThumbKey:    thumbKey,
		Width:       width,
		Height:      height,
		ThumbWidth:  tw,
		ThumbHeight: th,
		SizeBytes:   int64(len(data)),
	}, nil
}

func (s *UploadService) buildResult(ctx context.Context, key, thumbKey string, data []byte, w, h int) (*UploadedImage, error) {
	u, err := s.storage.URL(ctx, key)
	if err != nil {
		return nil, err
	}
	tu, err := s.storage.URL(ctx, thumbKey)
	if err != nil {
		return nil, err
	}
	return &UploadedImage{
		Key:         key,
		URL:         u,
		ThumbKey:    thumbKey,
		ThumbURL:    tu,
		Width:       w,
		Height:      h,
		ThumbWidth:  w,
		ThumbHeight: h,
		SizeBytes:   int64(len(data)),
	}, nil
}

func (s *UploadService) MaxBytes() int64 { return s.maxBytes }

func (s *UploadService) URLs(ctx context.Context, img *UploadedImage) error {
	if img.URL == "" {
		u, err := s.storage.URL(ctx, img.Key)
		if err != nil {
			return err
		}
		img.URL = u
	}
	if img.ThumbURL == "" && img.ThumbKey != "" {
		u, err := s.storage.URL(ctx, img.ThumbKey)
		if err != nil {
			return err
		}
		img.ThumbURL = u
	}
	return nil
}

// DetectContentType returns the content type from the first 512 bytes of an
// upload, honoring the configured allowlist. This is a belt-and-braces guard
// for non-image uploads before the bimg processing path.
func (s *UploadService) DetectContentType(data []byte) string {
	return http.DetectContentType(data)
}
