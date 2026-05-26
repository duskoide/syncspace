package api

import (
	"encoding/json"
	"net/http"

	"syncspace/backend/internal/models"
)

// ==================== Workspace Handlers ====================

func (h *Handler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r.Context())
	workspaces, err := h.svc.WorkspaceService.ListWorkspaces(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, workspaces)
}

func (h *Handler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req models.CreateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}

	claims := GetUserFromContext(r.Context())
	workspace, err := h.svc.WorkspaceService.CreateWorkspace(r.Context(), claims.UserID, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, workspace)
}

func (h *Handler) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/workspaces/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid workspace id")
		return
	}

	claims := GetUserFromContext(r.Context())
	workspace, err := h.svc.WorkspaceService.GetWorkspace(r.Context(), claims.UserID, id)
	if err != nil {
		writeError(w, 404, "not_found", err.Error())
		return
	}
	writeJSON(w, 200, workspace)
}

func (h *Handler) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/workspaces/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid workspace id")
		return
	}

	var req models.UpdateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}

	claims := GetUserFromContext(r.Context())
	workspace, err := h.svc.WorkspaceService.UpdateWorkspace(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, workspace)
}

func (h *Handler) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/workspaces/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid workspace id")
		return
	}

	claims := GetUserFromContext(r.Context())
	if err := h.svc.WorkspaceService.DeleteWorkspace(r.Context(), claims.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

// ==================== Note Handlers ====================

func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := parseID(r.URL.Path, "/api/workspaces/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid workspace id")
		return
	}

	claims := GetUserFromContext(r.Context())
	notes, err := h.svc.NoteService.ListNotesByWorkspace(r.Context(), claims.UserID, workspaceID)
	if err != nil {
		writeError(w, 404, "not_found", err.Error())
		return
	}
	writeJSON(w, 200, notes)
}

func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := parseID(r.URL.Path, "/api/workspaces/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid workspace id")
		return
	}

	var req models.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	req.WorkspaceID = workspaceID

	claims := GetUserFromContext(r.Context())
	note, err := h.svc.NoteService.CreateNote(r.Context(), claims.UserID, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, note)
}

func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/notes/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid note id")
		return
	}

	claims := GetUserFromContext(r.Context())
	note, err := h.svc.NoteService.GetNote(r.Context(), claims.UserID, id)
	if err != nil {
		writeError(w, 404, "not_found", err.Error())
		return
	}
	writeJSON(w, 200, note)
}

func (h *Handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/notes/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid note id")
		return
	}

	var req models.UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}

	claims := GetUserFromContext(r.Context())
	note, err := h.svc.NoteService.UpdateNote(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, note)
}

func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/notes/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid note id")
		return
	}

	claims := GetUserFromContext(r.Context())
	if err := h.svc.NoteService.DeleteNote(r.Context(), claims.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}
