package handlers

import (
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/kosero/atchannel/backend/pkg/storage"
)

// FileHandler serves objects from the Storage backend for the local storage
// provider. When S3 is used, files are served by the S3 bucket/CDN instead and
// this handler is not mounted.
type FileHandler struct {
	st storage.Storage
}

func NewFileHandler(st storage.Storage) *FileHandler {
	return &FileHandler{st: st}
}

func (h *FileHandler) Serve(w http.ResponseWriter, r *http.Request) {
	// chi URL param "key" may be empty when nothing matches; fall back to the
	// cleaned request path so nested keys like images/123.png still work.
	key := URLParam(r, "key")
	if key == "" {
		key = strings.TrimPrefix(r.URL.Path, "/files/")
	}
	key = path.Clean(key)
	if key == "." || strings.HasPrefix(key, "/") || strings.HasPrefix(key, "..") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	rc, err := h.st.Get(r.Context(), key)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rc.Close()

	w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	contentType := detectByExtension(key)
	w.Header().Set("Content-Type", contentType)
	_, _ = io.Copy(w, rc)
}

func detectByExtension(key string) string {
	switch path.Ext(key) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}
