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
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /api/tasks", h.listTasks)
	mux.HandleFunc("POST /api/tasks", h.createTask)
	mux.HandleFunc("GET /api/tasks/", h.getTask)
	mux.HandleFunc("PUT /api/tasks/", h.updateTask)
	mux.HandleFunc("DELETE /api/tasks/", h.deleteTask)
	mux.HandleFunc("GET /api/notes", h.listNotes)
	mux.HandleFunc("POST /api/notes", h.createNote)
	mux.HandleFunc("GET /api/notes/", h.getNote)
	mux.HandleFunc("PUT /api/notes/", h.updateNote)
	mux.HandleFunc("DELETE /api/notes/", h.deleteNote)
	mux.HandleFunc("POST /api/notes/", h.enrichNote)
	mux.HandleFunc("GET /api/wiki/summary", h.wikiSummary)
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
