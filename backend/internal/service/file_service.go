package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"syncspace/backend/internal/models"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

var allowedMimeTypes = map[string]bool{
	"image/jpeg":       true,
	"image/png":        true,
	"image/gif":        true,
	"video/mp4":        true,
	"video/webm":       true,
	"application/pdf":  true,
	"text/plain":       true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true,
	"application/vnd.ms-powerpoint": true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
}

func (s *Service) UploadFile(ctx context.Context, uploadedBy int64, boardID *int64, file multipart.File, header *multipart.FileHeader) (models.Attachment, error) {
	if header.Size > maxFileSize {
		return models.Attachment{}, fmt.Errorf("file size exceeds 10MB limit")
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if !allowedMimeTypes[mimeType] {
		return models.Attachment{}, fmt.Errorf("file type not allowed")
	}

	// Create upload directory
	uploadDir := filepath.Join("uploads", fmt.Sprintf("user_%d", uploadedBy))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return models.Attachment{}, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".bin"
	}
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	out, err := os.Create(filePath)
	if err != nil {
		return models.Attachment{}, fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return models.Attachment{}, fmt.Errorf("failed to save file: %w", err)
	}

	att := models.Attachment{
		BoardID:      boardID,
		Filename:     filename,
		OriginalName: header.Filename,
		MimeType:     mimeType,
		FileSize:     header.Size,
		FilePath:     filePath,
		UploadedBy:   uploadedBy,
	}

	return s.store.CreateAttachment(ctx, att)
}

func (s *Service) GetAttachment(ctx context.Context, id int64) (models.Attachment, error) {
	return s.store.GetAttachment(ctx, id)
}

func (s *Service) ListAttachmentsByBoard(ctx context.Context, boardID int64) ([]models.Attachment, error) {
	return s.store.ListAttachmentsByBoard(ctx, boardID)
}

func (s *Service) DeleteAttachment(ctx context.Context, userID, id int64) error {
	att, err := s.store.GetAttachment(ctx, id)
	if err != nil {
		return err
	}
	if att.UploadedBy != userID {
		return fmt.Errorf("not authorized")
	}
	
	// Delete physical file
	if att.FilePath != "" {
		_ = os.Remove(att.FilePath)
	}
	
	return s.store.DeleteAttachment(ctx, id)
}

func GetMimeTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return "application/octet-stream"
	}
}
