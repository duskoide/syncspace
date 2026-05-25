package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"syncspace/backend/internal/models"
)

func (h *Handler) createBoard(w http.ResponseWriter, r *http.Request) {
	var req models.Board
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.CreateBoard(r.Context(), claims.UserID, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) listBoards(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.ListBoards(r.Context(), claims.UserID, claims.Role)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) getBoard(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/boards/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	out, err := h.svc.GetBoard(r.Context(), id)
	if err != nil {
		writeError(w, 404, "not_found", "board not found")
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) updateBoard(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/boards/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	var req models.Board
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.UpdateBoard(r.Context(), claims.UserID, id, claims.Role, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) deleteBoard(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/boards/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.DeleteBoard(r.Context(), claims.UserID, id, claims.Role); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) joinBoard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	boardID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid board id")
		return
	}
	
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Role = "viewer" // default role
	}
	
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.JoinBoard(r.Context(), claims.UserID, boardID, req.Role)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) leaveBoard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	boardID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid board id")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.LeaveBoard(r.Context(), claims.UserID, boardID); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) updateMemberRole(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	membershipID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid membership id")
		return
	}
	
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	
	claims := GetUserFromContext(r.Context())
	if err := h.svc.UpdateMemberRole(r.Context(), claims.UserID, membershipID, req.Role); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "updated"})
}

func (h *Handler) listBoardMembers(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	boardID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid board id")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.ListBoardMemberships(r.Context(), claims.UserID, boardID)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) removeMember(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	boardID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid board id")
		return
	}
	memberIDStr := r.PathValue("member_id")
	memberID, err := strconv.ParseInt(memberIDStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid member id")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.RemoveMember(r.Context(), claims.UserID, boardID, memberID); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

// TextElement handlers

func (h *Handler) createTextElement(w http.ResponseWriter, r *http.Request) {
	var req models.TextElement
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.CreateTextElement(r.Context(), claims.UserID, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) listTextElements(w http.ResponseWriter, r *http.Request) {
	boardIDStr := r.URL.Query().Get("board_id")
	if boardIDStr == "" {
		writeError(w, 400, "bad_request", "board_id is required")
		return
	}
	boardID, err := strconv.ParseInt(boardIDStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid board_id")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.ListTextElementsByBoard(r.Context(), claims.UserID, boardID)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) getTextElement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/text-elements/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.GetTextElement(r.Context(), claims.UserID, id)
	if err != nil {
		writeError(w, 404, "not_found", "text element not found")
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) updateTextElement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/text-elements/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	var req models.TextElement
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.UpdateTextElement(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) deleteTextElement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/text-elements/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.DeleteTextElement(r.Context(), claims.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

// Discussion handlers

func (h *Handler) createDiscussion(w http.ResponseWriter, r *http.Request) {
	var req models.Discussion
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.CreateDiscussion(r.Context(), claims.UserID, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) listDiscussions(w http.ResponseWriter, r *http.Request) {
	boardIDStr := r.URL.Query().Get("board_id")
	if boardIDStr == "" {
		writeError(w, 400, "bad_request", "board_id is required")
		return
	}
	boardID, err := strconv.ParseInt(boardIDStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid board_id")
		return
	}
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.ListDiscussionsByBoard(r.Context(), claims.UserID, boardID, limit, offset)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) listDiscussionReplies(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	parentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid discussion id")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.ListDiscussionReplies(r.Context(), claims.UserID, parentID)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) deleteDiscussion(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/discussions/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.DeleteDiscussion(r.Context(), claims.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}
