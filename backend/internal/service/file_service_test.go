package service

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"syncspace/backend/internal/models"
)

// ==================== GetMimeTypeFromFilename ====================

func TestGetMimeTypeFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"photo.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"image.png", "image/png"},
		{"anim.gif", "image/gif"},
		{"pic.webp", "image/webp"},
		{"file.PDF", "application/octet-stream"},
		{"file.txt", "application/octet-stream"},
		{"file", "application/octet-stream"},
		{".jpg", "image/jpeg"},
		{".png", "image/png"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := GetMimeTypeFromFilename(tt.filename)
			if got != tt.expected {
				t.Errorf("GetMimeTypeFromFilename(%q) = %q, want %q", tt.filename, got, tt.expected)
			}
		})
	}
}

// ==================== Upload Note Image ====================

func TestUploadNoteImageSuccess(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "upload1@example.com", "password123", "Upload1", "user")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatal(err)
	}
	// Write minimal JPEG data (SOI marker)
	part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00})
	writer.WriteField("note_id", "0")
	writer.Close()

	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, fileHeader, err := req.FormFile("file")
	if err != nil {
		t.Fatal(err)
	}
	// The multipart form doesn't set Content-Type per part; we need to set it explicitly
	fileHeader.Header.Set("Content-Type", "image/jpeg")

	file, err := fileHeader.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	ni, err := svc.FileService.UploadNoteImage(context.Background(), user.ID, nil, file, fileHeader)
	if err != nil {
		t.Fatal(err)
	}
	if ni.OriginalName != "test.jpg" {
		t.Fatalf("expected original name 'test.jpg', got %s", ni.OriginalName)
	}
	if ni.MimeType != "image/jpeg" {
		t.Fatalf("expected mime type 'image/jpeg', got %s", ni.MimeType)
	}
	if ni.UploadedBy != user.ID {
		t.Fatalf("expected uploaded_by %d, got %d", user.ID, ni.UploadedBy)
	}
}

func TestUploadNoteImageInvalidType(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "upload2@example.com", "password123", "Upload2", "user")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "malware.exe")
	part.Write([]byte{0x00, 0x00, 0x00, 0x00})
	writer.Close()

	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, fileHeader, _ := req.FormFile("file")
	file, _ := fileHeader.Open()
	defer file.Close()

	_, err := svc.FileService.UploadNoteImage(context.Background(), user.ID, nil, file, fileHeader)
	if err == nil {
		t.Fatal("expected error for invalid file type")
	}
}

func TestUploadNoteImageTooLarge(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "upload3@example.com", "password123", "Upload3", "user")

	fh := &multipart.FileHeader{
		Filename: "huge.jpg",
		Size:     11 * 1024 * 1024, // 11MB
		Header:   map[string][]string{"Content-Type": {"image/jpeg"}},
	}

	_, err := svc.FileService.UploadNoteImage(context.Background(), user.ID, nil, nil, fh)
	if err == nil {
		t.Fatal("expected error for file too large")
	}
}

// ==================== Delete Note Image ====================

func TestDeleteNoteImageSuccess(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "delimg1@example.com", "password123", "DelImg1", "user")

	uploadDir := t.TempDir()
	os.MkdirAll(filepath.Join(uploadDir, "user_2"), 0755)
	testFile := filepath.Join(uploadDir, "user_2", "test.jpg")
	os.WriteFile(testFile, []byte("fake image"), 0644)

	ni, err := svc.store.CreateNoteImage(context.Background(), models.NoteImage{
		Filename:     "test.jpg",
		OriginalName: "test.jpg",
		MimeType:     "image/jpeg",
		FileSize:     10,
		FilePath:     testFile,
		UploadedBy:   user.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = svc.FileService.DeleteNoteImage(context.Background(), user.ID, ni.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(testFile)
	if os.IsExist(err) {
		t.Fatal("expected file to be deleted")
	}
}

func TestDeleteNoteImageAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "di2@example.com", "password123", "DI2", "user")
	user2 := registerUser(t, svc, "di3@example.com", "password123", "DI3", "user")

	ni, err := svc.store.CreateNoteImage(context.Background(), models.NoteImage{
		Filename:     "protected.jpg",
		OriginalName: "protected.jpg",
		MimeType:     "image/jpeg",
		FileSize:     10,
		UploadedBy:   user1.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = svc.FileService.DeleteNoteImage(context.Background(), user2.ID, ni.ID)
	if err == nil {
		t.Fatal("expected access denied for non-uploader deleting image")
	}
}

func TestDeleteNoteImageNotFound(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "di4@example.com", "password123", "DI4", "user")

	err := svc.FileService.DeleteNoteImage(context.Background(), user.ID, 999)
	if err == nil {
		t.Fatal("expected error for non-existent image")
	}
}

// ==================== GetNoteImage ====================

func TestGetNoteImage(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "gi1@example.com", "password123", "GI1", "user")

	ni, err := svc.store.CreateNoteImage(context.Background(), models.NoteImage{
		Filename:     "getme.jpg",
		OriginalName: "getme.jpg",
		MimeType:     "image/jpeg",
		FileSize:     100,
		UploadedBy:   user.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := svc.FileService.GetNoteImage(context.Background(), ni.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.OriginalName != "getme.jpg" {
		t.Fatalf("expected original name 'getme.jpg', got %s", got.OriginalName)
	}
}

func TestGetNoteImageNotFound(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.FileService.GetNoteImage(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent image")
	}
}
