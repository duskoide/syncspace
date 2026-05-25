package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"syncspace/backend/internal/models"
)

func (h *Handler) createMaterial(w http.ResponseWriter, r *http.Request) {
	var req models.Material
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.CreateMaterial(r.Context(), claims.UserID, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) listMaterials(w http.ResponseWriter, r *http.Request) {
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
	out, err := h.svc.ListMaterialsByClassroom(r.Context(), classroomID)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) getMaterial(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/materials/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	out, err := h.svc.GetMaterial(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, 404, "not_found", "material not found")
			return
		}
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) updateMaterial(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/materials/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	var req models.Material
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.UpdateMaterial(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) deleteMaterial(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/materials/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.DeleteMaterial(r.Context(), claims.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) createAssignment(w http.ResponseWriter, r *http.Request) {
	var req models.Assignment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.CreateAssignment(r.Context(), claims.UserID, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) listAssignments(w http.ResponseWriter, r *http.Request) {
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
	out, err := h.svc.ListAssignmentsByClassroom(r.Context(), classroomID)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) getAssignment(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/assignments/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	out, err := h.svc.GetAssignment(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, 404, "not_found", "assignment not found")
			return
		}
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) updateAssignment(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/assignments/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	var req models.Assignment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.UpdateAssignment(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) deleteAssignment(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/assignments/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.DeleteAssignment(r.Context(), claims.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) createSubmission(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	assignmentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid assignment id")
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.CreateSubmission(r.Context(), claims.UserID, assignmentID, req.Content)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) listSubmissions(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	assignmentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid assignment id")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.ListSubmissionsByAssignment(r.Context(), claims.UserID, assignmentID)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) gradeSubmission(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	submissionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid submission id")
		return
	}
	var req struct {
		Score    int    `json:"score"`
		Feedback string `json:"feedback"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.GradeSubmission(r.Context(), claims.UserID, submissionID, req.Score, req.Feedback); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "graded"})
}
