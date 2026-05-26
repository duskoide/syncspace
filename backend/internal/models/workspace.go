package models

import "time"

// Workspace represents a user's personal note-taking container
type Workspace struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UserID      int64     `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Note represents a rich text note within a workspace
type Note struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"` // HTML content from TipTap editor
	CreatedBy   int64     `json:"created_by"`
	CreatorName string    `json:"creator_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Template represents a shared workspace or note that can be cloned
type Template struct {
	ID              int64     `json:"id"`
	Type            string    `json:"type"`             // "workspace" or "note"
	SourceID        int64     `json:"source_id"`        // workspace_id or note_id
	CreatorID       int64     `json:"creator_id"`
	CreatorName     string    `json:"creator_name,omitempty"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Visibility      string    `json:"visibility"`       // "public" or "link"
	ContentSnapshot string    `json:"content_snapshot"` // JSON string
	IsHidden        bool      `json:"is_hidden"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// NoteImage represents an inline image uploaded to a note
type NoteImage struct {
	ID           int64     `json:"id"`
	NoteID       int64     `json:"note_id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int64     `json:"file_size"`
	FilePath     string    `json:"-"`
	UploadedBy   int64     `json:"uploaded_by"`
	UserName     string    `json:"user_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// TemplateSnapshot is used to store workspace/note content as JSON
type TemplateSnapshot struct {
	WorkspaceID int64        `json:"workspace_id,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Notes       []Note       `json:"notes,omitempty"`
	Note        *Note        `json:"note,omitempty"`
}

// CreateWorkspaceRequest represents a request to create a workspace
type CreateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateWorkspaceRequest represents a request to update a workspace
type UpdateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateNoteRequest represents a request to create a note
type CreateNoteRequest struct {
	WorkspaceID int64  `json:"workspace_id"`
	Title       string `json:"title"`
}

// UpdateNoteRequest represents a request to update a note
type UpdateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// CreateTemplateRequest represents a request to create a template
type CreateTemplateRequest struct {
	Type        string `json:"type"`        // "workspace" or "note"
	SourceID    int64  `json:"source_id"`   // workspace_id or note_id
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`  // "public" or "link"
}

// UpdateTemplateRequest represents a request to update template metadata
type UpdateTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

// CloneTemplateRequest represents a request to clone a template
type CloneTemplateRequest struct {
	TargetWorkspaceID int64 `json:"target_workspace_id"` // For note templates, where to create the note
}
