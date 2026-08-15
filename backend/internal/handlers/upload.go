package handlers

import (
	"io"
	"net/http"

	"github.com/kosero/atchannel/backend/internal/services"
)

type UploadHandler struct {
	svc *services.UploadService
}

func NewUploadHandler(svc *services.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(int64(h.svc.MaxBytes()) + 1<<20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, h.svc.MaxBytes()+1))
	if err != nil {
		writeErr(w, err)
		return
	}

	img, err := h.svc.Store(r.Context(), data, h.svc.DetectContentType(data))
	if err != nil {
		writeErr(w, err)
		return
	}
	if err := h.svc.URLs(r.Context(), img); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, img)
}
