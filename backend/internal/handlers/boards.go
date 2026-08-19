package handlers

import (
	"net/http"
	"strconv"

	"github.com/kosero/atchannel/backend/internal/services"
)

type BoardHandler struct {
	svc *services.BoardService
}

func NewBoardHandler(svc *services.BoardService) *BoardHandler {
	return &BoardHandler{svc: svc}
}

func (h *BoardHandler) List(w http.ResponseWriter, r *http.Request) {
	boards, err := h.svc.List(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"boards": boards})
}

type createBoardRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *BoardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createBoardRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	board, err := h.svc.Create(r.Context(), req.Slug, req.Name, req.Description)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, board)
}

type ThreadHandler struct {
	svc   *services.ThreadService
	users *services.UserService
}

func NewThreadHandler(svc *services.ThreadService, users *services.UserService) *ThreadHandler {
	return &ThreadHandler{svc: svc, users: users}
}

type createThreadRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	ImageKey string `json:"imageKey"`
}

func (h *ThreadHandler) ListByBoard(w http.ResponseWriter, r *http.Request) {
	slug := URLParam(r, "board")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))
	threads, total, err := h.svc.ListByBoard(r.Context(), slug, page, perPage)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"threads": threads,
		"total":   total,
	})
}

func (h *ThreadHandler) Create(w http.ResponseWriter, r *http.Request) {
	slug := URLParam(r, "board")
	var req createThreadRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.users.CurrentUser(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	thread, err := h.svc.Create(r.Context(), user, slug, req.Title, req.Body, req.ImageKey)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, thread)
}

func (h *ThreadHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := URLParam(r, "id")
	thread, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	replies, err := h.svc.ListReplies(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"thread":  thread,
		"replies": replies,
	})
}

type createReplyRequest struct {
	Body      string  `json:"body"`
	ImageKey  string  `json:"imageKey"`
	ReplyToID *string `json:"replyToId"`
}

func (h *ThreadHandler) Reply(w http.ResponseWriter, r *http.Request) {
	id := URLParam(r, "id")
	var req createReplyRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.users.CurrentUser(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	reply, err := h.svc.Reply(r.Context(), user, id, req.Body, req.ImageKey, req.ReplyToID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, reply)
}
