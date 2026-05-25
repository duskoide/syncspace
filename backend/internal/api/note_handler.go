package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"syncspace/backend/internal/models"
)

func (h *Handler) createCollaborativeNote(w http.ResponseWriter, r *http.Request) {
	var req models.CollaborativeNote
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.CreateCollaborativeNote(r.Context(), claims.UserID, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) listCollaborativeNotes(w http.ResponseWriter, r *http.Request) {
	classroomIDStr := r.URL.Query().Get("classroom_id")
	if classroomIDStr == "" {
		writeError(w, 400, "bad_request", "classroom_id is required")
		return
	}
	classroomID, err := strconv.ParseInt(classroomIDStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid classroom_id")
		return
	}
	out, err := h.svc.ListCollaborativeNotesByClassroom(r.Context(), classroomID)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) getCollaborativeNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/collaborative-notes/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	out, err := h.svc.GetCollaborativeNote(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, 404, "not_found", "note not found")
			return
		}
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) updateCollaborativeNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/collaborative-notes/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	var req models.CollaborativeNote
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.UpdateCollaborativeNote(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) deleteCollaborativeNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/collaborative-notes/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.DeleteCollaborativeNote(r.Context(), claims.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

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
	classroomIDStr := r.URL.Query().Get("classroom_id")
	if classroomIDStr == "" {
		writeError(w, 400, "bad_request", "classroom_id is required")
		return
	}
	classroomID, err := strconv.ParseInt(classroomIDStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid classroom_id")
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
	out, err := h.svc.ListDiscussionsByClassroom(r.Context(), classroomID, limit, offset)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
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
	out, err := h.svc.ListDiscussionReplies(r.Context(), parentID)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}
