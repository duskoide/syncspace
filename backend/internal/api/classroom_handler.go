package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"syncspace/backend/internal/models"
)

func (h *Handler) createClassroom(w http.ResponseWriter, r *http.Request) {
	var req models.Classroom
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.CreateClassroom(r.Context(), claims.UserID, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) listClassrooms(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.ListClassrooms(r.Context(), claims.UserID, claims.Role)
	if err != nil {
		writeError(w, 500, "internal_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) getClassroom(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/classrooms/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	out, err := h.svc.GetClassroom(r.Context(), id)
	if err != nil {
		writeError(w, 404, "not_found", "classroom not found")
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) updateClassroom(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/classrooms/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	var req models.Classroom
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", "invalid json")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.UpdateClassroom(r.Context(), claims.UserID, id, req)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) deleteClassroom(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/classrooms/")
	if !ok {
		writeError(w, 400, "bad_request", "invalid id")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.DeleteClassroom(r.Context(), claims.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) requestEnrollment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	classroomID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid classroom id")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.RequestEnrollment(r.Context(), claims.UserID, classroomID)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 201, out)
}

func (h *Handler) approveEnrollment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	enrollmentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid enrollment id")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.ApproveEnrollment(r.Context(), claims.UserID, enrollmentID); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "approved"})
}

func (h *Handler) removeStudent(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	classroomID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid classroom id")
		return
	}
	studentIDStr := r.PathValue("student_id")
	studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid student id")
		return
	}
	claims := GetUserFromContext(r.Context())
	if err := h.svc.RemoveStudent(r.Context(), claims.UserID, classroomID, studentID); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) listClassroomStudents(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	classroomID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid classroom id")
		return
	}
	claims := GetUserFromContext(r.Context())
	out, err := h.svc.ListEnrollmentsByClassroom(r.Context(), claims.UserID, classroomID, "")
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	writeJSON(w, 200, out)
}
