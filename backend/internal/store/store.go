package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"syncspace/backend/internal/auth"
	"syncspace/backend/internal/models"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("enable wal: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		return nil, fmt.Errorf("busy timeout: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := s.createDefaultSuperadmin(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create default superadmin: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate(ctx context.Context) error {
	// Drop all old tables from previous schema
	dropOldTables := `
DROP TABLE IF EXISTS collaborative_notes;
DROP TABLE IF EXISTS note_contributors;
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS assignments;
DROP TABLE IF EXISTS materials;
DROP TABLE IF EXISTS attachments_old;
DROP TABLE IF EXISTS enrollments;
DROP TABLE IF EXISTS classrooms;
DROP TABLE IF EXISTS boards;
DROP TABLE IF EXISTS board_memberships;
DROP TABLE IF EXISTS text_elements;
DROP TABLE IF EXISTS discussions;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS notes;
`
	_, _ = s.db.ExecContext(ctx, dropOldTables)

	schema := `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	name TEXT NOT NULL,
	role TEXT NOT NULL CHECK(role IN ('superadmin', 'creator', 'user')),
	status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'active', 'suspended')),
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

CREATE TABLE IF NOT EXISTS workspaces (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_workspaces_user ON workspaces(user_id);

CREATE TABLE IF NOT EXISTS notes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	content TEXT NOT NULL DEFAULT '',
	created_by INTEGER NOT NULL REFERENCES users(id),
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_notes_workspace ON notes(workspace_id);
CREATE INDEX IF NOT EXISTS idx_notes_created_by ON notes(created_by);

CREATE TABLE IF NOT EXISTS templates (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	type TEXT NOT NULL CHECK(type IN ('workspace', 'note')),
	source_id INTEGER NOT NULL,
	creator_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	visibility TEXT NOT NULL CHECK(visibility IN ('public', 'link')) DEFAULT 'public',
	content_snapshot TEXT NOT NULL,
	is_hidden BOOLEAN DEFAULT FALSE,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_templates_creator ON templates(creator_id);
CREATE INDEX IF NOT EXISTS idx_templates_visibility ON templates(visibility);
CREATE INDEX IF NOT EXISTS idx_templates_type ON templates(type);

CREATE TABLE IF NOT EXISTS note_images (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	note_id INTEGER REFERENCES notes(id) ON DELETE SET NULL,
	filename TEXT NOT NULL,
	original_name TEXT NOT NULL,
	mime_type TEXT NOT NULL,
	file_size INTEGER NOT NULL,
	file_path TEXT NOT NULL,
	uploaded_by INTEGER NOT NULL REFERENCES users(id),
	created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_note_images_note ON note_images(note_id);
CREATE INDEX IF NOT EXISTS idx_note_images_uploaded_by ON note_images(uploaded_by);
`
	_, err := s.db.ExecContext(ctx, schema)
	return err
}

func (s *Store) createDefaultSuperadmin(ctx context.Context) error {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = 'superadmin'`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := auth.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("hash default admin password: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO users(email, password_hash, name, role, status, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?)`,
		"admin@syncspace.edu", hash, "System Admin", "superadmin", "active", now, now)
	return err
}

// ==================== User Methods ====================

func (s *Store) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var u models.User
	var c, up string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, role, status, created_at, updated_at FROM users WHERE email = ?`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.Status, &c, &up)
	if err != nil {
		return u, err
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, c)
	u.UpdatedAt, _ = time.Parse(time.RFC3339, up)
	return u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	var u models.User
	var c, up string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, role, status, created_at, updated_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.Status, &c, &up)
	if err != nil {
		return u, err
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, c)
	u.UpdatedAt, _ = time.Parse(time.RFC3339, up)
	return u, nil
}

func (s *Store) CreateUser(ctx context.Context, u models.User) (models.User, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users(email, password_hash, name, role, status, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?)`,
		u.Email, u.PasswordHash, u.Name, u.Role, u.Status, now, now)
	if err != nil {
		return models.User{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetUserByID(ctx, id)
}

func (s *Store) UpdateUserStatus(ctx context.Context, id int64, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET status = ?, updated_at = ? WHERE id = ?`,
		status, now, id)
	return err
}

func (s *Store) ListUsers(ctx context.Context, role, status string) ([]models.User, error) {
	query := `SELECT id, email, name, role, status, created_at, updated_at FROM users WHERE 1=1`
	args := []interface{}{}
	if role != "" {
		query += ` AND role = ?`
		args = append(args, role)
	}
	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.User{}
	for rows.Next() {
		var u models.User
		var c, up string
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Status, &c, &up); err != nil {
			return nil, err
		}
		u.CreatedAt, _ = time.Parse(time.RFC3339, c)
		u.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Store) DeleteUser(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}

// ==================== Workspace Methods ====================

func (s *Store) CreateWorkspace(ctx context.Context, w models.Workspace) (models.Workspace, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO workspaces(name, description, user_id, created_at, updated_at) VALUES(?, ?, ?, ?, ?)`,
		w.Name, w.Description, w.UserID, now, now)
	if err != nil {
		return models.Workspace{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetWorkspace(ctx, id)
}

func (s *Store) GetWorkspace(ctx context.Context, id int64) (models.Workspace, error) {
	var w models.Workspace
	var c, up string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, user_id, created_at, updated_at FROM workspaces WHERE id = ?`, id).
		Scan(&w.ID, &w.Name, &w.Description, &w.UserID, &c, &up)
	if err != nil {
		return w, err
	}
	w.CreatedAt, _ = time.Parse(time.RFC3339, c)
	w.UpdatedAt, _ = time.Parse(time.RFC3339, up)
	return w, nil
}

func (s *Store) ListWorkspacesByUser(ctx context.Context, userID int64) ([]models.Workspace, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, description, user_id, created_at, updated_at FROM workspaces WHERE user_id = ? ORDER BY updated_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Workspace{}
	for rows.Next() {
		var w models.Workspace
		var c, up string
		if err := rows.Scan(&w.ID, &w.Name, &w.Description, &w.UserID, &c, &up); err != nil {
			return nil, err
		}
		w.CreatedAt, _ = time.Parse(time.RFC3339, c)
		w.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *Store) UpdateWorkspace(ctx context.Context, id int64, w models.Workspace) (models.Workspace, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`UPDATE workspaces SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		w.Name, w.Description, now, id)
	if err != nil {
		return models.Workspace{}, err
	}
	return s.GetWorkspace(ctx, id)
}

func (s *Store) DeleteWorkspace(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM workspaces WHERE id = ?`, id)
	return err
}

// ==================== Note Methods ====================

func (s *Store) CreateNote(ctx context.Context, n models.Note) (models.Note, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO notes(workspace_id, title, content, created_by, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?)`,
		n.WorkspaceID, n.Title, n.Content, n.CreatedBy, now, now)
	if err != nil {
		return models.Note{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetNote(ctx, id)
}

func (s *Store) GetNote(ctx context.Context, id int64) (models.Note, error) {
	var n models.Note
	var c, up string
	err := s.db.QueryRowContext(ctx,
		`SELECT n.id, n.workspace_id, n.title, n.content, n.created_by, u.name, n.created_at, n.updated_at 
		 FROM notes n JOIN users u ON n.created_by = u.id WHERE n.id = ?`, id).
		Scan(&n.ID, &n.WorkspaceID, &n.Title, &n.Content, &n.CreatedBy, &n.CreatorName, &c, &up)
	if err != nil {
		return n, err
	}
	n.CreatedAt, _ = time.Parse(time.RFC3339, c)
	n.UpdatedAt, _ = time.Parse(time.RFC3339, up)
	return n, nil
}

func (s *Store) ListNotesByWorkspace(ctx context.Context, workspaceID int64) ([]models.Note, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT n.id, n.workspace_id, n.title, n.content, n.created_by, u.name, n.created_at, n.updated_at 
		 FROM notes n JOIN users u ON n.created_by = u.id 
		 WHERE n.workspace_id = ? ORDER BY n.updated_at DESC`,
		workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Note{}
	for rows.Next() {
		var n models.Note
		var c, up string
		if err := rows.Scan(&n.ID, &n.WorkspaceID, &n.Title, &n.Content, &n.CreatedBy, &n.CreatorName, &c, &up); err != nil {
			return nil, err
		}
		n.CreatedAt, _ = time.Parse(time.RFC3339, c)
		n.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Store) UpdateNote(ctx context.Context, id int64, n models.Note) (models.Note, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`UPDATE notes SET title = ?, content = ?, updated_at = ? WHERE id = ?`,
		n.Title, n.Content, now, id)
	if err != nil {
		return models.Note{}, err
	}
	return s.GetNote(ctx, id)
}

func (s *Store) DeleteNote(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ?`, id)
	return err
}

// ==================== Template Methods ====================

func (s *Store) CreateTemplate(ctx context.Context, t models.Template) (models.Template, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO templates(type, source_id, creator_id, name, description, visibility, content_snapshot, is_hidden, created_at, updated_at) 
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Type, t.SourceID, t.CreatorID, t.Name, t.Description, t.Visibility, t.ContentSnapshot, t.IsHidden, now, now)
	if err != nil {
		return models.Template{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetTemplate(ctx, id)
}

func (s *Store) GetTemplate(ctx context.Context, id int64) (models.Template, error) {
	var t models.Template
	var c, up string
	err := s.db.QueryRowContext(ctx,
		`SELECT t.id, t.type, t.source_id, t.creator_id, u.name, t.name, t.description, t.visibility, t.content_snapshot, t.is_hidden, t.created_at, t.updated_at 
		 FROM templates t JOIN users u ON t.creator_id = u.id WHERE t.id = ?`, id).
		Scan(&t.ID, &t.Type, &t.SourceID, &t.CreatorID, &t.CreatorName, &t.Name, &t.Description, &t.Visibility, &t.ContentSnapshot, &t.IsHidden, &c, &up)
	if err != nil {
		return t, err
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, c)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, up)
	return t, nil
}

func (s *Store) ListTemplates(ctx context.Context, visibility string, search string, excludeHidden bool) ([]models.Template, error) {
	query := `SELECT t.id, t.type, t.source_id, t.creator_id, u.name, t.name, t.description, t.visibility, t.content_snapshot, t.is_hidden, t.created_at, t.updated_at 
			  FROM templates t JOIN users u ON t.creator_id = u.id WHERE 1=1`
	args := []interface{}{}
	
	if visibility != "" {
		query += ` AND t.visibility = ?`
		args = append(args, visibility)
	}
	if search != "" {
		query += ` AND (t.name LIKE ? OR t.description LIKE ?)`
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	if excludeHidden {
		query += ` AND t.is_hidden = FALSE`
	}
	query += ` ORDER BY t.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Template{}
	for rows.Next() {
		var t models.Template
		var c, up string
		if err := rows.Scan(&t.ID, &t.Type, &t.SourceID, &t.CreatorID, &t.CreatorName, &t.Name, &t.Description, &t.Visibility, &t.ContentSnapshot, &t.IsHidden, &c, &up); err != nil {
			return nil, err
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, c)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) ListTemplatesByCreator(ctx context.Context, creatorID int64) ([]models.Template, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT t.id, t.type, t.source_id, t.creator_id, u.name, t.name, t.description, t.visibility, t.content_snapshot, t.is_hidden, t.created_at, t.updated_at 
		 FROM templates t JOIN users u ON t.creator_id = u.id WHERE t.creator_id = ? ORDER BY t.updated_at DESC`,
		creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Template{}
	for rows.Next() {
		var t models.Template
		var c, up string
		if err := rows.Scan(&t.ID, &t.Type, &t.SourceID, &t.CreatorID, &t.CreatorName, &t.Name, &t.Description, &t.Visibility, &t.ContentSnapshot, &t.IsHidden, &c, &up); err != nil {
			return nil, err
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, c)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) UpdateTemplate(ctx context.Context, id int64, t models.Template) (models.Template, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`UPDATE templates SET name = ?, description = ?, visibility = ?, content_snapshot = ?, updated_at = ? WHERE id = ?`,
		t.Name, t.Description, t.Visibility, t.ContentSnapshot, now, id)
	if err != nil {
		return models.Template{}, err
	}
	return s.GetTemplate(ctx, id)
}

func (s *Store) UpdateTemplateHidden(ctx context.Context, id int64, isHidden bool) error {
	_, err := s.db.ExecContext(ctx, `UPDATE templates SET is_hidden = ? WHERE id = ?`, isHidden, id)
	return err
}

func (s *Store) DeleteTemplate(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM templates WHERE id = ?`, id)
	return err
}

// ==================== NoteImage Methods ====================

func (s *Store) CreateNoteImage(ctx context.Context, ni models.NoteImage) (models.NoteImage, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO note_images(note_id, filename, original_name, mime_type, file_size, file_path, uploaded_by, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		ni.NoteID, ni.Filename, ni.OriginalName, ni.MimeType, ni.FileSize, ni.FilePath, ni.UploadedBy, now)
	if err != nil {
		return models.NoteImage{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetNoteImage(ctx, id)
}

func (s *Store) GetNoteImage(ctx context.Context, id int64) (models.NoteImage, error) {
	var ni models.NoteImage
	var c string
	var nid sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT ni.id, ni.note_id, ni.filename, ni.original_name, ni.mime_type, ni.file_size, ni.file_path, ni.uploaded_by, u.name, ni.created_at 
		 FROM note_images ni JOIN users u ON ni.uploaded_by = u.id WHERE ni.id = ?`, id).
		Scan(&ni.ID, &nid, &ni.Filename, &ni.OriginalName, &ni.MimeType, &ni.FileSize, &ni.FilePath, &ni.UploadedBy, &ni.UserName, &c)
	if err != nil {
		return ni, err
	}
	if nid.Valid {
		ni.NoteID = nid.Int64
	}
	ni.CreatedAt, _ = time.Parse(time.RFC3339, c)
	return ni, nil
}

func (s *Store) ListNoteImagesByNote(ctx context.Context, noteID int64) ([]models.NoteImage, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT ni.id, ni.note_id, ni.filename, ni.original_name, ni.mime_type, ni.file_size, ni.file_path, ni.uploaded_by, u.name, ni.created_at 
		 FROM note_images ni JOIN users u ON ni.uploaded_by = u.id 
		 WHERE ni.note_id = ? ORDER BY ni.created_at DESC`,
		noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.NoteImage{}
	for rows.Next() {
		var ni models.NoteImage
		var c string
		var nid sql.NullInt64
		if err := rows.Scan(&ni.ID, &nid, &ni.Filename, &ni.OriginalName, &ni.MimeType, &ni.FileSize, &ni.FilePath, &ni.UploadedBy, &ni.UserName, &c); err != nil {
			return nil, err
		}
		if nid.Valid {
			ni.NoteID = nid.Int64
		}
		ni.CreatedAt, _ = time.Parse(time.RFC3339, c)
		out = append(out, ni)
	}
	return out, rows.Err()
}

func (s *Store) DeleteNoteImage(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM note_images WHERE id = ?`, id)
	return err
}

// ==================== JSON Helper ====================

func ToJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func FromJSON(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}
