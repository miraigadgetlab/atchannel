package handlers

import (
	"net/http"
	"net/netip"
	"strconv"
	"time"

	"github.com/kosero/atchannel/backend/internal/services"
)

type AdminHandler struct {
	svc   *services.ModerationService
	users *services.UserService
}

func NewAdminHandler(svc *services.ModerationService, users *services.UserService) *AdminHandler {
	return &AdminHandler{svc: svc, users: users}
}

func (h *AdminHandler) DeleteThread(w http.ResponseWriter, r *http.Request) {
	id := URLParam(r, "id")
	if err := h.svc.DeleteThread(r.Context(), id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) DeleteReply(w http.ResponseWriter, r *http.Request) {
	id := URLParam(r, "id")
	if err := h.svc.DeleteReply(r.Context(), id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))
	reports, err := h.svc.ListReports(r.Context(), status, page, perPage)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reports": reports})
}

type resolveReportRequest struct {
	Status string `json:"status"`
}

func (h *AdminHandler) ResolveReport(w http.ResponseWriter, r *http.Request) {
	id := URLParam(r, "id")
	var req resolveReportRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	report, err := h.svc.ResolveReport(r.Context(), id, req.Status)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

type createBanRequest struct {
	UserID    *string `json:"userId"`
	IP        string  `json:"ip"`
	Reason    string  `json:"reason"`
	ExpiresIn string  `json:"expiresIn"`
}

func (h *AdminHandler) CreateBan(w http.ResponseWriter, r *http.Request) {
	var req createBanRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == nil && req.IP == "" {
		writeError(w, http.StatusBadRequest, "userId or ip is required")
		return
	}
	var expiresAt *time.Time
	if req.ExpiresIn != "" {
		d, err := time.ParseDuration(req.ExpiresIn)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid expiresIn duration")
			return
		}
		t := time.Now().Add(d)
		expiresAt = &t
	}

	if req.UserID != nil {
		ban, err := h.svc.BanUser(r.Context(), *req.UserID, req.Reason, expiresAt)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, ban)
		return
	}

	ip, err := netip.ParseAddr(req.IP)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ip address")
		return
	}
	ban, err := h.svc.BanIP(r.Context(), ip, req.Reason, expiresAt)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ban)
}

func (h *AdminHandler) ListBans(w http.ResponseWriter, r *http.Request) {
	bans, err := h.svc.ListBans(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bans": bans})
}
