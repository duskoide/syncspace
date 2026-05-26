package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"syncspace/backend/internal/models"
	"syncspace/backend/internal/store"
)

// FileService handles file uploads and downloads for note images
type FileService struct {
	store     *store.Store
	uploadDir string
}

// Allowed MIME types for note images
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

func NewFileService(store *store.Store, uploadDir string) *FileService {
	return &FileService{
		store:     store,
		uploadDir: uploadDir,
	}
}

// UploadNoteImage uploads an image for a note
func (s *FileService) UploadNoteImage(ctx context.Context, uploadedBy int64, noteID int64, file multipart.File, header *multipart.FileHeader) (models.NoteImage, error) {
	// Validate file size (max 10MB)
	const maxSize = 10 * 1024 * 1024
	if header.Size > maxSize {
		return models.NoteImage{}, errors.New("file too large (max 10MB)")
	}

	// Validate MIME type
	mimeType := header.Header.Get("Content-Type")
	if !allowedImageTypes[mimeType] {
		return models.NoteImage{}, fmt.Errorf("invalid file type: %s (allowed: jpeg, png, gif, webp)", mimeType)
	}

	// Create upload directory
	userDir := filepath.Join(s.uploadDir, fmt.Sprintf("user_%d", uploadedBy))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return models.NoteImage{}, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(userDir, filename)

	// Save file
	out, err := os.Create(filePath)
	if err != nil {
		return models.NoteImage{}, fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		os.Remove(filePath)
		return models.NoteImage{}, fmt.Errorf("failed to save file: %w", err)
	}

	// Save to database
	ni := models.NoteImage{
		NoteID:       noteID,
		Filename:     filename,
		OriginalName: header.Filename,
		MimeType:     mimeType,
		FileSize:     header.Size,
		FilePath:     filePath,
		UploadedBy:   uploadedBy,
	}

	return s.store.CreateNoteImage(ctx, ni)
}

// GetNoteImage retrieves a note image by ID
func (s *FileService) GetNoteImage(ctx context.Context, id int64) (models.NoteImage, error) {
	return s.store.GetNoteImage(ctx, id)
}

// DeleteNoteImage deletes a note image
func (s *FileService) DeleteNoteImage(ctx context.Context, userID int64, imageID int64) error {
	ni, err := s.store.GetNoteImage(ctx, imageID)
	if err != nil {
		return err
	}

	// Only uploader can delete
	if ni.UploadedBy != userID {
		return errors.New("access denied")
	}

	// Delete file
	if ni.FilePath != "" {
		os.Remove(ni.FilePath)
	}

	// Delete from database
	return s.store.DeleteNoteImage(ctx, imageID)
}

// GetMimeTypeFromFilename returns MIME type based on file extension
func GetMimeTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
