package handlers

import (
	"net/http"

	"github.com/kosero/atchannel/backend/internal/services"
)

type ReportHandler struct {
	svc   *services.ModerationService
	users *services.UserService
}

func NewReportHandler(svc *services.ModerationService, users *services.UserService) *ReportHandler {
	return &ReportHandler{svc: svc, users: users}
}

type createReportRequest struct {
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Reason     string `json:"reason"`
}

func (h *ReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, err := h.users.CurrentUser(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	var req createReportRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	report, err := h.svc.CreateReport(r.Context(), user.ID, req.TargetType, req.TargetID, req.Reason)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, report)
}
