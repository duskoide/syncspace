package api

import (
	"encoding/json"
	"net/http"

	"syncspace/backend/internal/models"
)

// ==================== Template Handlers ====================

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	templates, err := h.svc.TemplateService.ListTemplates(r.Context(), search)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, templates)
}

func (h *Handler) ListMyTemplates(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r.Context())

	templates, err := h.svc.TemplateService.ListMyTemplates(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, templates)
}

func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/templates/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid template id")
		return
	}

	claims := GetUserFromContext(r.Context())
	template, err := h.svc.TemplateService.GetTemplate(r.Context(), claims.UserID, id)
	if err != nil {
		writeError(w, 404, "not_found", err.Error())
		return
	}
	writeJSON(w, 200, template)
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}

	claims := GetUserFromContext(r.Context())
	template, err := h.svc.TemplateService.CreateTemplate(r.Context(), claims.UserID, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, template)
}

func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/templates/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid template id")
		return
	}

	var req models.UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}

	claims := GetUserFromContext(r.Context())
	template, err := h.svc.TemplateService.UpdateTemplate(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, template)
}

func (h *Handler) UpdateTemplateContent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/templates/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid template id")
		return
	}

	claims := GetUserFromContext(r.Context())
	template, err := h.svc.TemplateService.UpdateTemplateContent(r.Context(), claims.UserID, id)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, template)
}

func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/templates/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid template id")
		return
	}

	claims := GetUserFromContext(r.Context())
	if err := h.svc.TemplateService.DeleteTemplate(r.Context(), claims.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) CloneTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/templates/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid template id")
		return
	}

	var req models.CloneTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}

	claims := GetUserFromContext(r.Context())
	workspace, note, err := h.svc.TemplateService.CloneTemplate(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}

	if workspace != nil {
		writeJSON(w, 201, map[string]interface{}{
			"type":      "workspace",
			"workspace": workspace,
		})
	} else {
		writeJSON(w, 201, map[string]interface{}{
			"type": "note",
			"note": note,
		})
	}
}

// ==================== Admin Template Handlers ====================

func (h *Handler) ListAllTemplatesAdmin(w http.ResponseWriter, r *http.Request) {
	templates, err := h.svc.TemplateService.ListAllTemplatesForAdmin(r.Context())
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, templates)
}

func (h *Handler) SetTemplateHidden(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/admin/templates/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid template id")
		return
	}

	var req struct {
		IsHidden bool `json:"is_hidden"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}

	claims := GetUserFromContext(r.Context())
	if err := h.svc.TemplateService.SetTemplateHidden(r.Context(), claims.UserID, id, req.IsHidden); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "updated"})
}
