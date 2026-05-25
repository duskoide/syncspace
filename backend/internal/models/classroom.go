package models

import "time"

type Classroom struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TeacherID   int64     `json:"teacher_id"`
	TeacherName string    `json:"teacher_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Enrollment struct {
	ID           int64     `json:"id"`
	ClassroomID  int64     `json:"classroom_id"`
	StudentID    int64     `json:"student_id"`
	StudentName  string    `json:"student_name,omitempty"`
	StudentEmail string    `json:"student_email,omitempty"`
	Status       string    `json:"status"` // pending, active, inactive
	EnrolledAt   time.Time `json:"enrolled_at"`
}

type Material struct {
	ID          int64     `json:"id"`
	ClassroomID int64     `json:"classroom_id"`
	TeacherID   int64     `json:"teacher_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Tags        string    `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Attachment struct {
	ID           int64     `json:"id"`
	MaterialID   *int64    `json:"material_id,omitempty"`
	SubmissionID *int64    `json:"submission_id,omitempty"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int64     `json:"file_size"`
	FilePath     string    `json:"-"`
	UploadedBy   int64     `json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
}

type Assignment struct {
	ID          int64     `json:"id"`
	ClassroomID int64     `json:"classroom_id"`
	TeacherID   int64     `json:"teacher_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	MaxScore    int       `json:"max_score"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Submission struct {
	ID           int64     `json:"id"`
	AssignmentID int64     `json:"assignment_id"`
	StudentID    int64     `json:"student_id"`
	StudentName  string    `json:"student_name,omitempty"`
	Content      string    `json:"content"`
	Score        *int      `json:"score,omitempty"`
	Feedback     string    `json:"feedback,omitempty"`
	SubmittedAt  time.Time `json:"submitted_at"`
	GradedAt     *time.Time `json:"graded_at,omitempty"`
}

type CollaborativeNote struct {
	ID          int64     `json:"id"`
	MaterialID  *int64    `json:"material_id,omitempty"`
	ClassroomID int64     `json:"classroom_id"`
	CreatedBy   int64     `json:"created_by"`
	CreatorName string    `json:"creator_name,omitempty"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Discussion struct {
	ID          int64     `json:"id"`
	ClassroomID int64     `json:"classroom_id"`
	MaterialID  *int64    `json:"material_id,omitempty"`
	UserID      int64     `json:"user_id"`
	UserName    string    `json:"user_name,omitempty"`
	Message     string    `json:"message"`
	ParentID    *int64    `json:"parent_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
