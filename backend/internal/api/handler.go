package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"syncspace/backend/internal/service"
)

// IsUpstreamError is a wrapper to access service.IsUpstreamError
func IsUpstreamError(err error) bool {
	return service.IsUpstreamError(err)
}

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

	// Workspace routes
	authMux.HandleFunc("GET /api/workspaces", h.ListWorkspaces)
	authMux.HandleFunc("POST /api/workspaces", h.CreateWorkspace)
	authMux.HandleFunc("GET /api/workspaces/", h.handleWorkspace)
	authMux.HandleFunc("PUT /api/workspaces/", h.handleWorkspace)
	authMux.HandleFunc("DELETE /api/workspaces/", h.handleWorkspace)

	// Note routes (nested under workspaces and standalone)
	authMux.HandleFunc("GET /api/workspaces/{id}/notes", h.ListNotes)
	authMux.HandleFunc("POST /api/workspaces/{id}/notes", h.CreateNote)
	authMux.HandleFunc("GET /api/notes/", h.handleNote)
	authMux.HandleFunc("PUT /api/notes/", h.handleNote)
	authMux.HandleFunc("DELETE /api/notes/", h.handleNote)

	// Template routes
	authMux.HandleFunc("GET /api/templates", h.ListTemplates)
	authMux.HandleFunc("GET /api/templates/my", h.ListMyTemplates)
	authMux.HandleFunc("POST /api/templates", h.CreateTemplate)
	authMux.HandleFunc("GET /api/templates/", h.handleTemplate)
	authMux.HandleFunc("PUT /api/templates/", h.handleTemplate)
	authMux.HandleFunc("DELETE /api/templates/", h.handleTemplate)
	authMux.HandleFunc("POST /api/templates/{id}/clone", h.CloneTemplate)
	authMux.HandleFunc("POST /api/templates/{id}/update-content", h.UpdateTemplateContent)

	// Admin template routes
	authMux.HandleFunc("GET /api/admin/templates", h.ListAllTemplatesAdmin)
	authMux.HandleFunc("PATCH /api/admin/templates/", h.SetTemplateHidden)

	// File upload/download for note images
	authMux.HandleFunc("POST /api/upload", h.uploadNoteImage)
	authMux.HandleFunc("GET /api/files/{id}", h.downloadFile)
	authMux.HandleFunc("DELETE /api/files/{id}", h.deleteNoteImage)

	// Wikipedia integration
	authMux.HandleFunc("GET /api/wiki/summary", h.wikiSummary)

	// Wrap auth routes with auth middleware
	mux.Handle("/api/auth/me", AuthMiddleware(authMux))
	mux.Handle("/api/admin/", AuthMiddleware(RequireRole("superadmin")(authMux)))
	mux.Handle("/api/workspaces", AuthMiddleware(authMux))
	mux.Handle("/api/workspaces/", AuthMiddleware(authMux))
	mux.Handle("/api/notes/", AuthMiddleware(authMux))
	mux.Handle("/api/templates", AuthMiddleware(authMux))
	mux.Handle("/api/templates/", AuthMiddleware(authMux))
	mux.Handle("/api/upload", AuthMiddleware(authMux))
	mux.Handle("/api/files/", AuthMiddleware(authMux))
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
		if strings.Contains(r.URL.Path, "/activate") {
			h.ActivateUser(w, r)
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

func (h *Handler) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetWorkspace(w, r)
	case http.MethodPut:
		h.UpdateWorkspace(w, r)
	case http.MethodDelete:
		h.DeleteWorkspace(w, r)
	default:
		writeError(w, 405, "method_not_allowed", "method not allowed")
	}
}

func (h *Handler) handleNote(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetNote(w, r)
	case http.MethodPut:
		h.UpdateNote(w, r)
	case http.MethodDelete:
		h.DeleteNote(w, r)
	default:
		writeError(w, 405, "method_not_allowed", "method not allowed")
	}
}

func (h *Handler) handleTemplate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetTemplate(w, r)
	case http.MethodPut:
		h.UpdateTemplate(w, r)
	case http.MethodDelete:
		h.DeleteTemplate(w, r)
	default:
		writeError(w, 405, "method_not_allowed", "method not allowed")
	}
}

func (h *Handler) wikiSummary(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	summary, err := h.svc.WikiSummary(r.Context(), topic)
	if err != nil {
		if IsUpstreamError(err) {
			writeError(w, 502, "upstream_error", err.Error())
			return
		}
		writeError(w, 400, "bad_request", err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"topic": topic, "summary": summary})
}
