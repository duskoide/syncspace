package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func (h *Handler) uploadNoteImage(w http.ResponseWriter, r *http.Request) {
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

	// Get note_id
	noteIDStr := r.FormValue("note_id")
	if noteIDStr == "" {
		writeError(w, 400, "bad_request", "note_id is required")
		return
	}
	noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid note_id")
		return
	}

	claims := GetUserFromContext(r.Context())
	image, err := h.svc.FileService.UploadNoteImage(r.Context(), claims.UserID, noteID, file, header)
	if err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}

	writeJSON(w, 201, image)
}

func (h *Handler) downloadFile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid file id")
		return
	}

	image, err := h.svc.FileService.GetNoteImage(r.Context(), id)
	if err != nil {
		writeError(w, 404, "not_found", "file not found")
		return
	}

	// Serve file
	data, err := os.ReadFile(image.FilePath)
	if err != nil {
		writeError(w, 500, "internal_error", "failed to read file")
		return
	}

	w.Header().Set("Content-Type", image.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, image.OriginalName))
	w.Header().Set("Content-Length", strconv.FormatInt(image.FileSize, 10))
	w.WriteHeader(200)
	w.Write(data)
}

func (h *Handler) deleteNoteImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "bad_request", "invalid image id")
		return
	}

	claims := GetUserFromContext(r.Context())
	if err := h.svc.FileService.DeleteNoteImage(r.Context(), claims.UserID, id); err != nil {
		writeError(w, 400, "validation_error", err.Error())
		return
	}
	w.WriteHeader(204)
}
