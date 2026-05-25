package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, 400, "bad_request", "failed to parse form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "bad_request", "file is required")
		return
	}
	defer file.Close()

	claims := GetUserFromContext(r.Context())
	att, err := h.svc.UploadFile(r.Context(), claims.UserID, file, header)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}

	writeJSON(w, 201, att)
}

func (h *Handler) downloadFile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid file id")
		return
	}

	att, err := h.svc.GetAttachment(r.Context(), id)
	if err != nil {
		writeError(w, 404, "not_found", "file not found")
		return
	}

	// Serve file
	data, err := os.ReadFile(att.FilePath)
	if err != nil {
		writeError(w, 500, "internal_error", "failed to read file")
		return
	}

	w.Header().Set("Content-Type", att.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.OriginalName))
	w.Header().Set("Content-Length", strconv.FormatInt(att.FileSize, 10))
	w.WriteHeader(200)
	w.Write(data)
}
