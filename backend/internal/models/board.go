package models

import "time"

type Board struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	ModeratorID   int64     `json:"moderator_id"`
	ModeratorName string    `json:"moderator_name,omitempty"`
	Visibility    string    `json:"visibility"` // public, private
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BoardMembership struct {
	ID        int64     `json:"id"`
	BoardID   int64     `json:"board_id"`
	UserID    int64     `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"`
	Role      string    `json:"role"` // viewer, editor
	JoinedAt  time.Time `json:"joined_at"`
}

type TextElement struct {
	ID          int64     `json:"id"`
	BoardID     int64     `json:"board_id"`
	Content     string    `json:"content"`
	X           float64   `json:"x"`
	Y           float64   `json:"y"`
	Width       float64   `json:"width"`
	Height      float64   `json:"height"`
	Color       string    `json:"color"`
	CreatedBy   int64     `json:"created_by"`
	CreatorName string    `json:"creator_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Discussion struct {
	ID       int64     `json:"id"`
	BoardID  int64     `json:"board_id"`
	UserID   int64     `json:"user_id"`
	UserName string    `json:"user_name,omitempty"`
	Message  string    `json:"message"`
	ParentID *int64    `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Attachment struct {
	ID           int64     `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int64     `json:"file_size"`
	FilePath     string    `json:"-"`
	UploadedBy   int64     `json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
}
