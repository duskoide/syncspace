package api

import (
	"encoding/json"
	"net/http"

	"syncspace/backend/internal/models"
)

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}

	u, err := h.svc.Register(r.Context(), req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}

	// Don't return password hash
	u.PasswordHash = ""
	writeJSON(w, 201, u)
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}

	token, u, err := h.svc.Login(r.Context(), req)
	if err != nil {
		writeError(w, 401, "unauthorized", err.Error())
		return
	}

	writeJSON(w, 200, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":         u.ID,
			"email":      u.Email,
			"name":       u.Name,
			"role":       u.Role,
			"status":     u.Status,
			"created_at": u.CreatedAt,
		},
	})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, 401, "unauthorized", "not authenticated")
		return
	}

	u, err := h.svc.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}

	u.PasswordHash = ""
	writeJSON(w, 200, u)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")

	users, err := h.svc.ListUsers(r.Context(), role, status)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}

	// Don't return password hashes
	for i := range users {
		users[i].PasswordHash = ""
	}
	writeJSON(w, 200, users)
}

func (h *Handler) ApproveUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/admin/users/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid user id")
		return
	}

	admin := GetUserFromContext(r.Context())
	if err := h.svc.ApproveUser(r.Context(), admin.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "approved"})
}

func (h *Handler) SuspendUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/admin/users/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid user id")
		return
	}

	admin := GetUserFromContext(r.Context())
	if err := h.svc.SuspendUser(r.Context(), admin.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "suspended"})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/admin/users/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid user id")
		return
	}

	admin := GetUserFromContext(r.Context())
	if err := h.svc.DeleteUser(r.Context(), admin.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}
