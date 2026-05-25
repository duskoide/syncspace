package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"syncspace/backend/internal/models"
	"syncspace/backend/internal/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler { return &Handler{svc: svc} }

type errResp struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	var e errResp
	e.Error.Code = code
	e.Error.Message = msg
	writeJSON(w, status, e)
}

func (h *Handler) Register(mux *http.ServeMux) {
	// Public routes
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /api/auth/register", h.HandleRegister)
	mux.HandleFunc("POST /api/auth/login", h.HandleLogin)

	// Auth middleware wrapper
	authMux := http.NewServeMux()

	// Authenticated routes
	authMux.HandleFunc("GET /api/auth/me", h.GetMe)

	// Admin routes
	authMux.HandleFunc("GET /api/admin/users", h.ListUsers)
	authMux.HandleFunc("PUT /api/admin/users/", h.handleAdminUser)
	authMux.HandleFunc("DELETE /api/admin/users/", h.handleAdminUser)

	// Classroom routes
	authMux.HandleFunc("POST /api/classrooms", h.createClassroom)
	authMux.HandleFunc("GET /api/classrooms", h.listClassrooms)
	authMux.HandleFunc("GET /api/classrooms/", h.handleClassroom)
	authMux.HandleFunc("PUT /api/classrooms/", h.handleClassroom)
	authMux.HandleFunc("DELETE /api/classrooms/", h.handleClassroom)

	// Enrollment routes
	authMux.HandleFunc("POST /api/classrooms/{id}/enroll", h.requestEnrollment)
	authMux.HandleFunc("PUT /api/enrollments/{id}/approve", h.approveEnrollment)
	authMux.HandleFunc("GET /api/classrooms/{id}/students", h.listClassroomStudents)
	authMux.HandleFunc("DELETE /api/classrooms/{id}/students/{student_id}", h.removeStudent)

	// Material routes
	authMux.HandleFunc("POST /api/materials", h.createMaterial)
	authMux.HandleFunc("GET /api/materials", h.listMaterials)
	authMux.HandleFunc("GET /api/materials/", h.handleMaterial)
	authMux.HandleFunc("PUT /api/materials/", h.handleMaterial)
	authMux.HandleFunc("DELETE /api/materials/", h.handleMaterial)

	// Assignment routes
	authMux.HandleFunc("POST /api/assignments", h.createAssignment)
	authMux.HandleFunc("GET /api/assignments", h.listAssignments)
	authMux.HandleFunc("GET /api/assignments/", h.handleAssignment)
	authMux.HandleFunc("PUT /api/assignments/", h.handleAssignment)
	authMux.HandleFunc("DELETE /api/assignments/", h.handleAssignment)
	authMux.HandleFunc("POST /api/assignments/{id}/submissions", h.createSubmission)
	authMux.HandleFunc("GET /api/assignments/{id}/submissions", h.listSubmissions)
	authMux.HandleFunc("PUT /api/submissions/{id}/grade", h.gradeSubmission)

	// Collaborative Note routes
	authMux.HandleFunc("POST /api/collaborative-notes", h.createCollaborativeNote)
	authMux.HandleFunc("GET /api/collaborative-notes", h.listCollaborativeNotes)
	authMux.HandleFunc("GET /api/collaborative-notes/", h.handleCollaborativeNote)
	authMux.HandleFunc("PUT /api/collaborative-notes/", h.handleCollaborativeNote)
	authMux.HandleFunc("DELETE /api/collaborative-notes/", h.handleCollaborativeNote)

	// Discussion routes
	authMux.HandleFunc("POST /api/discussions", h.createDiscussion)
	authMux.HandleFunc("GET /api/discussions", h.listDiscussions)
	authMux.HandleFunc("GET /api/discussions/{id}/replies", h.listDiscussionReplies)

	// File upload/download
	authMux.HandleFunc("POST /api/upload", h.uploadFile)
	authMux.HandleFunc("GET /api/files/{id}", h.downloadFile)

	// Legacy routes (keep for backward compatibility)
	authMux.HandleFunc("GET /api/tasks", h.listTasks)
	authMux.HandleFunc("POST /api/tasks", h.createTask)
	authMux.HandleFunc("GET /api/tasks/", h.getTask)
	authMux.HandleFunc("PUT /api/tasks/", h.updateTask)
	authMux.HandleFunc("DELETE /api/tasks/", h.deleteTask)
	authMux.HandleFunc("GET /api/notes", h.listNotes)
	authMux.HandleFunc("POST /api/notes", h.createNote)
	authMux.HandleFunc("GET /api/notes/", h.getNote)
	authMux.HandleFunc("PUT /api/notes/", h.updateNote)
	authMux.HandleFunc("DELETE /api/notes/", h.deleteNote)
	authMux.HandleFunc("POST /api/notes/", h.enrichNote)
	authMux.HandleFunc("GET /api/wiki/summary", h.wikiSummary)

	// Wrap auth routes with auth middleware
	mux.Handle("/api/auth/me", AuthMiddleware(authMux))
	mux.Handle("/api/admin/", AuthMiddleware(RequireRole("superadmin")(authMux)))
	mux.Handle("/api/classrooms", AuthMiddleware(authMux))
	mux.Handle("/api/classrooms/", AuthMiddleware(authMux))
	mux.Handle("/api/enrollments/", AuthMiddleware(authMux))
	mux.Handle("/api/materials", AuthMiddleware(authMux))
	mux.Handle("/api/materials/", AuthMiddleware(authMux))
	mux.Handle("/api/assignments", AuthMiddleware(authMux))
	mux.Handle("/api/assignments/", AuthMiddleware(authMux))
	mux.Handle("/api/submissions/", AuthMiddleware(authMux))
	mux.Handle("/api/collaborative-notes", AuthMiddleware(authMux))
	mux.Handle("/api/collaborative-notes/", AuthMiddleware(authMux))
	mux.Handle("/api/discussions", AuthMiddleware(authMux))
	mux.Handle("/api/discussions/", AuthMiddleware(authMux))
	mux.Handle("/api/upload", AuthMiddleware(authMux))
	mux.Handle("/api/files/", AuthMiddleware(authMux))
	mux.Handle("/api/tasks", AuthMiddleware(authMux))
	mux.Handle("/api/tasks/", AuthMiddleware(authMux))
	mux.Handle("/api/notes", AuthMiddleware(authMux))
	mux.Handle("/api/notes/", AuthMiddleware(authMux))
	mux.Handle("/api/wiki/", AuthMiddleware(authMux))
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func parseID(path, prefix string) (int64, bool) {
	raw := strings.TrimPrefix(path, prefix)
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return 0, false
	}
	parts := strings.Split(raw, "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (h *Handler) handleAdminUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		if strings.Contains(r.URL.Path, "/approve") {
			h.ApproveUser(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "/suspend") {
			h.SuspendUser(w, r)
			return
		}
		writeError(w, 400, "bad_request", "unknown admin action")
	case http.MethodDelete:
		h.DeleteUser(w, r)
	default:
		writeError(w, 405, "method_not_allowed", "method not allowed")
	}
}

func (h *Handler) handleClassroom(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getClassroom(w, r)
	case http.MethodPut:
		h.updateClassroom(w, r)
	case http.MethodDelete:
		h.deleteClassroom(w, r)
	default:
		writeError(w, 405, "method_not_allowed", "method not allowed")
	}
}

func (h *Handler) handleMaterial(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getMaterial(w, r)
	case http.MethodPut:
		h.updateMaterial(w, r)
	case http.MethodDelete:
		h.deleteMaterial(w, r)
	default:
		writeError(w, 405, "method_not_allowed", "method not allowed")
	}
}

func (h *Handler) handleAssignment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getAssignment(w, r)
	case http.MethodPut:
		h.updateAssignment(w, r)
	case http.MethodDelete:
		h.deleteAssignment(w, r)
	default:
		writeError(w, 405, "method_not_allowed", "method not allowed")
	}
}

func (h *Handler) handleCollaborativeNote(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getCollaborativeNote(w, r)
	case http.MethodPut:
		h.updateCollaborativeNote(w, r)
	case http.MethodDelete:
		h.deleteCollaborativeNote(w, r)
	default:
		writeError(w, 405, "method_not_allowed", "method not allowed")
	}
}

// Legacy handlers

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListTasks(r.Context())
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, items)
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var in models.Task
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	out, err := h.svc.CreateTask(r.Context(), in)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/tasks/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	out, err := h.svc.GetTask(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, 404, "not_found", "task not found")
			return
		}
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/tasks/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	var in models.Task
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	out, err := h.svc.UpdateTask(r.Context(), id, in)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, 404, "not_found", "task not found")
			return
		}
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/tasks/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	if err := h.svc.DeleteTask(r.Context(), id); err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) listNotes(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListNotes(r.Context())
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, items)
}

func (h *Handler) createNote(w http.ResponseWriter, r *http.Request) {
	var in models.Note
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	out, err := h.svc.CreateNote(r.Context(), in)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) getNote(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/enrich") {
		h.enrichNote(w, r)
		return
	}
	id, ok := parseID(r.URL.Path, "/api/notes/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	out, err := h.svc.GetNote(r.Context(), id)
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

func (h *Handler) updateNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/notes/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	var in models.Note
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	out, err := h.svc.UpdateNote(r.Context(), id, in)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, 404, "not_found", "note not found")
			return
		}
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/notes/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	if err := h.svc.DeleteNote(r.Context(), id); err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) wikiSummary(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	summary, err := h.svc.WikiSummary(r.Context(), topic)
	if err != nil {
		if service.IsUpstreamError(err) {
			writeError(w, 502, "upstream_error", err.Error())
			return
		}
		writeError(w, 400, "bad_request", err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"topic": topic, "summary": summary})
}

func (h *Handler) enrichNote(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/enrich") {
		writeError(w, 404, "not_found", "not found")
		return
	}
	raw := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/notes/"), "/enrich")
	raw = strings.Trim(raw, "/")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	var in struct {
		Topic string `json:"topic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	out, err := h.svc.EnrichNote(r.Context(), id, in.Topic)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, 404, "not_found", "note not found")
			return
		}
		if service.IsUpstreamError(err) {
			writeError(w, 502, "upstream_error", err.Error())
			return
		}
		writeError(w, 400, "bad_request", err.Error())
		return
	}
	writeJSON(w, 200, out)
}
