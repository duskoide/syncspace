package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

	// Create default superadmin if none exists
	if err := s.createDefaultSuperadmin(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create default superadmin: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate(ctx context.Context) error {
	// Drop old tables from previous schema to ensure clean migration
	dropOldTables := `
DROP TABLE IF EXISTS collaborative_notes;
DROP TABLE IF EXISTS note_contributors;
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS assignments;
DROP TABLE IF EXISTS materials;
DROP TABLE IF EXISTS attachments_old;
DROP TABLE IF EXISTS enrollments;
DROP TABLE IF EXISTS classrooms;
`
	_, _ = s.db.ExecContext(ctx, dropOldTables)

	schema := `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	name TEXT NOT NULL,
	role TEXT NOT NULL CHECK(role IN ('superadmin', 'moderator', 'collaborator')),
	status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'active', 'suspended')),
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

CREATE TABLE IF NOT EXISTS boards (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	moderator_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_boards_moderator ON boards(moderator_id);

CREATE TABLE IF NOT EXISTS board_memberships (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	board_id INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	role TEXT DEFAULT 'viewer' CHECK(role IN ('viewer', 'editor')),
	joined_at TEXT NOT NULL,
	UNIQUE(board_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_board_memberships_board ON board_memberships(board_id);
CREATE INDEX IF NOT EXISTS idx_board_memberships_user ON board_memberships(user_id);

CREATE TABLE IF NOT EXISTS text_elements (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	board_id INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
	content TEXT NOT NULL DEFAULT '',
	x REAL NOT NULL DEFAULT 0,
	y REAL NOT NULL DEFAULT 0,
	width REAL NOT NULL DEFAULT 200,
	height REAL NOT NULL DEFAULT 150,
	color TEXT NOT NULL DEFAULT '#FFFF88',
	created_by INTEGER NOT NULL REFERENCES users(id),
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_text_elements_board ON text_elements(board_id);
CREATE INDEX IF NOT EXISTS idx_text_elements_created_by ON text_elements(created_by);

CREATE TABLE IF NOT EXISTS discussions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	board_id INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	message TEXT NOT NULL,
	parent_id INTEGER REFERENCES discussions(id) ON DELETE CASCADE,
	created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_discussions_board ON discussions(board_id);

CREATE TABLE IF NOT EXISTS attachments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	filename TEXT NOT NULL,
	original_name TEXT NOT NULL,
	mime_type TEXT NOT NULL,
	file_size INTEGER NOT NULL,
	file_path TEXT NOT NULL,
	uploaded_by INTEGER NOT NULL REFERENCES users(id),
	created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_attachments_uploaded_by ON attachments(uploaded_by);

-- Legacy tables for backward compatibility (kept but not used)
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	due_date TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS notes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	content TEXT NOT NULL,
	tags TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);`
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

	// Create default superadmin: admin@syncspace.edu / admin123
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

// User methods

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

// Board methods

func (s *Store) CreateBoard(ctx context.Context, b models.Board) (models.Board, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO boards(name, description, moderator_id, created_at, updated_at) VALUES(?, ?, ?, ?, ?)`,
		b.Name, b.Description, b.ModeratorID, now, now)
	if err != nil {
		return models.Board{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetBoard(ctx, id)
}

func (s *Store) GetBoard(ctx context.Context, id int64) (models.Board, error) {
	var b models.Board
	var cr, up string
	err := s.db.QueryRowContext(ctx,
		`SELECT b.id, b.name, b.description, b.moderator_id, u.name, b.created_at, b.updated_at 
		 FROM boards b JOIN users u ON b.moderator_id = u.id WHERE b.id = ?`, id).
		Scan(&b.ID, &b.Name, &b.Description, &b.ModeratorID, &b.ModeratorName, &cr, &up)
	if err != nil {
		return b, err
	}
	b.CreatedAt, _ = time.Parse(time.RFC3339, cr)
	b.UpdatedAt, _ = time.Parse(time.RFC3339, up)
	return b, nil
}

func (s *Store) ListBoards(ctx context.Context, moderatorID int64) ([]models.Board, error) {
	query := `SELECT b.id, b.name, b.description, b.moderator_id, u.name, b.created_at, b.updated_at 
		      FROM boards b JOIN users u ON b.moderator_id = u.id`
	args := []interface{}{}
	if moderatorID > 0 {
		query += ` WHERE b.moderator_id = ?`
		args = append(args, moderatorID)
	}
	query += ` ORDER BY b.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Board{}
	for rows.Next() {
		var b models.Board
		var cr, up string
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.ModeratorID, &b.ModeratorName, &cr, &up); err != nil {
			return nil, err
		}
		b.CreatedAt, _ = time.Parse(time.RFC3339, cr)
		b.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) ListBoardsByMember(ctx context.Context, userID int64) ([]models.Board, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT b.id, b.name, b.description, b.moderator_id, u.name, b.created_at, b.updated_at 
		 FROM boards b 
		 JOIN users u ON b.moderator_id = u.id 
		 JOIN board_memberships bm ON b.id = bm.board_id 
		 WHERE bm.user_id = ?
		 ORDER BY b.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Board{}
	for rows.Next() {
		var b models.Board
		var cr, up string
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.ModeratorID, &b.ModeratorName, &cr, &up); err != nil {
			return nil, err
		}
		b.CreatedAt, _ = time.Parse(time.RFC3339, cr)
		b.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) UpdateBoard(ctx context.Context, id int64, b models.Board) (models.Board, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`UPDATE boards SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		b.Name, b.Description, now, id)
	if err != nil {
		return models.Board{}, err
	}
	return s.GetBoard(ctx, id)
}

func (s *Store) DeleteBoard(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM boards WHERE id = ?`, id)
	return err
}

// BoardMembership methods

func (s *Store) CreateBoardMembership(ctx context.Context, bm models.BoardMembership) (models.BoardMembership, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO board_memberships(board_id, user_id, role, joined_at) VALUES(?, ?, ?, ?)`,
		bm.BoardID, bm.UserID, bm.Role, now)
	if err != nil {
		return models.BoardMembership{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetBoardMembership(ctx, id)
}

func (s *Store) GetBoardMembership(ctx context.Context, id int64) (models.BoardMembership, error) {
	var bm models.BoardMembership
	var joined string
	err := s.db.QueryRowContext(ctx,
		`SELECT bm.id, bm.board_id, bm.user_id, u.name, bm.role, bm.joined_at
		 FROM board_memberships bm JOIN users u ON bm.user_id = u.id WHERE bm.id = ?`, id).
		Scan(&bm.ID, &bm.BoardID, &bm.UserID, &bm.UserName, &bm.Role, &joined)
	if err != nil {
		return bm, err
	}
	bm.JoinedAt, _ = time.Parse(time.RFC3339, joined)
	return bm, nil
}

func (s *Store) GetBoardMembershipByBoardAndUser(ctx context.Context, boardID, userID int64) (models.BoardMembership, error) {
	var bm models.BoardMembership
	var joined string
	err := s.db.QueryRowContext(ctx,
		`SELECT bm.id, bm.board_id, bm.user_id, u.name, bm.role, bm.joined_at
		 FROM board_memberships bm JOIN users u ON bm.user_id = u.id 
		 WHERE bm.board_id = ? AND bm.user_id = ?`, boardID, userID).
		Scan(&bm.ID, &bm.BoardID, &bm.UserID, &bm.UserName, &bm.Role, &joined)
	if err != nil {
		return bm, err
	}
	bm.JoinedAt, _ = time.Parse(time.RFC3339, joined)
	return bm, nil
}

func (s *Store) ListBoardMembershipsByBoard(ctx context.Context, boardID int64) ([]models.BoardMembership, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT bm.id, bm.board_id, bm.user_id, u.name, bm.role, bm.joined_at
		 FROM board_memberships bm JOIN users u ON bm.user_id = u.id 
		 WHERE bm.board_id = ? ORDER BY bm.joined_at DESC`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.BoardMembership{}
	for rows.Next() {
		var bm models.BoardMembership
		var joined string
		if err := rows.Scan(&bm.ID, &bm.BoardID, &bm.UserID, &bm.UserName, &bm.Role, &joined); err != nil {
			return nil, err
		}
		bm.JoinedAt, _ = time.Parse(time.RFC3339, joined)
		out = append(out, bm)
	}
	return out, rows.Err()
}

func (s *Store) UpdateBoardMembershipRole(ctx context.Context, id int64, role string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE board_memberships SET role = ? WHERE id = ?`,
		role, id)
	return err
}

func (s *Store) DeleteBoardMembership(ctx context.Context, boardID, userID int64) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM board_memberships WHERE board_id = ? AND user_id = ?`,
		boardID, userID)
	return err
}

// TextElement methods

func (s *Store) CreateTextElement(ctx context.Context, te models.TextElement) (models.TextElement, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO text_elements(board_id, content, x, y, width, height, color, created_by, created_at, updated_at) 
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		te.BoardID, te.Content, te.X, te.Y, te.Width, te.Height, te.Color, te.CreatedBy, now, now)
	if err != nil {
		return models.TextElement{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetTextElement(ctx, id)
}

func (s *Store) GetTextElement(ctx context.Context, id int64) (models.TextElement, error) {
	var te models.TextElement
	var c, up string
	err := s.db.QueryRowContext(ctx,
		`SELECT te.id, te.board_id, te.content, te.x, te.y, te.width, te.height, te.color, te.created_by, u.name, te.created_at, te.updated_at
		 FROM text_elements te JOIN users u ON te.created_by = u.id WHERE te.id = ?`, id).
		Scan(&te.ID, &te.BoardID, &te.Content, &te.X, &te.Y, &te.Width, &te.Height, &te.Color, &te.CreatedBy, &te.CreatorName, &c, &up)
	if err != nil {
		return te, err
	}
	te.CreatedAt, _ = time.Parse(time.RFC3339, c)
	te.UpdatedAt, _ = time.Parse(time.RFC3339, up)
	return te, nil
}

func (s *Store) ListTextElementsByBoard(ctx context.Context, boardID int64) ([]models.TextElement, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT te.id, te.board_id, te.content, te.x, te.y, te.width, te.height, te.color, te.created_by, u.name, te.created_at, te.updated_at
		 FROM text_elements te JOIN users u ON te.created_by = u.id 
		 WHERE te.board_id = ? ORDER BY te.updated_at DESC`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.TextElement{}
	for rows.Next() {
		var te models.TextElement
		var c, up string
		if err := rows.Scan(&te.ID, &te.BoardID, &te.Content, &te.X, &te.Y, &te.Width, &te.Height, &te.Color, &te.CreatedBy, &te.CreatorName, &c, &up); err != nil {
			return nil, err
		}
		te.CreatedAt, _ = time.Parse(time.RFC3339, c)
		te.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, te)
	}
	return out, rows.Err()
}

func (s *Store) UpdateTextElement(ctx context.Context, id int64, te models.TextElement) (models.TextElement, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`UPDATE text_elements SET content = ?, x = ?, y = ?, width = ?, height = ?, color = ?, updated_at = ? WHERE id = ?`,
		te.Content, te.X, te.Y, te.Width, te.Height, te.Color, now, id)
	if err != nil {
		return models.TextElement{}, err
	}
	return s.GetTextElement(ctx, id)
}

func (s *Store) DeleteTextElement(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM text_elements WHERE id = ?`, id)
	return err
}

// Discussion methods

func (s *Store) CreateDiscussion(ctx context.Context, d models.Discussion) (models.Discussion, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO discussions(board_id, user_id, message, parent_id, created_at) VALUES(?, ?, ?, ?, ?)`,
		d.BoardID, d.UserID, d.Message, d.ParentID, now)
	if err != nil {
		return models.Discussion{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetDiscussion(ctx, id)
}

func (s *Store) GetDiscussion(ctx context.Context, id int64) (models.Discussion, error) {
	var d models.Discussion
	var pid sql.NullInt64
	var c string
	err := s.db.QueryRowContext(ctx,
		`SELECT d.id, d.board_id, d.user_id, u.name, d.message, d.parent_id, d.created_at
		 FROM discussions d JOIN users u ON d.user_id = u.id WHERE d.id = ?`, id).
		Scan(&d.ID, &d.BoardID, &d.UserID, &d.UserName, &d.Message, &pid, &c)
	if err != nil {
		return d, err
	}
	if pid.Valid {
		d.ParentID = &pid.Int64
	}
	d.CreatedAt, _ = time.Parse(time.RFC3339, c)
	return d, nil
}

func (s *Store) ListDiscussionsByBoard(ctx context.Context, boardID int64, limit, offset int) ([]models.Discussion, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT d.id, d.board_id, d.user_id, u.name, d.message, d.parent_id, d.created_at
		 FROM discussions d JOIN users u ON d.user_id = u.id WHERE d.board_id = ? AND d.parent_id IS NULL
		 ORDER BY d.created_at DESC LIMIT ? OFFSET ?`, boardID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Discussion{}
	for rows.Next() {
		var d models.Discussion
		var pid sql.NullInt64
		var c string
		if err := rows.Scan(&d.ID, &d.BoardID, &d.UserID, &d.UserName, &d.Message, &pid, &c); err != nil {
			return nil, err
		}
		if pid.Valid {
			d.ParentID = &pid.Int64
		}
		d.CreatedAt, _ = time.Parse(time.RFC3339, c)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) ListDiscussionReplies(ctx context.Context, parentID int64) ([]models.Discussion, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT d.id, d.board_id, d.user_id, u.name, d.message, d.parent_id, d.created_at
		 FROM discussions d JOIN users u ON d.user_id = u.id WHERE d.parent_id = ? ORDER BY d.created_at ASC`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Discussion{}
	for rows.Next() {
		var d models.Discussion
		var pid sql.NullInt64
		var c string
		if err := rows.Scan(&d.ID, &d.BoardID, &d.UserID, &d.UserName, &d.Message, &pid, &c); err != nil {
			return nil, err
		}
		if pid.Valid {
			d.ParentID = &pid.Int64
		}
		d.CreatedAt, _ = time.Parse(time.RFC3339, c)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) DeleteDiscussion(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM discussions WHERE id = ?`, id)
	return err
}

// Attachment methods

func (s *Store) CreateAttachment(ctx context.Context, a models.Attachment) (models.Attachment, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO attachments(filename, original_name, mime_type, file_size, file_path, uploaded_by, created_at) VALUES(?, ?, ?, ?, ?, ?, ?)`,
		a.Filename, a.OriginalName, a.MimeType, a.FileSize, a.FilePath, a.UploadedBy, now)
	if err != nil {
		return models.Attachment{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetAttachment(ctx, id)
}

func (s *Store) GetAttachment(ctx context.Context, id int64) (models.Attachment, error) {
	var a models.Attachment
	var c string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, filename, original_name, mime_type, file_size, file_path, uploaded_by, created_at FROM attachments WHERE id = ?`, id).
		Scan(&a.ID, &a.Filename, &a.OriginalName, &a.MimeType, &a.FileSize, &a.FilePath, &a.UploadedBy, &c)
	if err != nil {
		return a, err
	}
	a.CreatedAt, _ = time.Parse(time.RFC3339, c)
	return a, nil
}

func (s *Store) DeleteAttachment(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM attachments WHERE id = ?`, id)
	return err
}

// Legacy methods (keep for backward compatibility)

func (s *Store) ListTasks(ctx context.Context) ([]models.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, description, status, due_date, created_at, updated_at FROM tasks ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Task{}
	for rows.Next() {
		var t models.Task
		var due sql.NullString
		var c, u string
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &due, &c, &u); err != nil {
			return nil, err
		}
		parseTimes(&t, due, c, u)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) GetTask(ctx context.Context, id int64) (models.Task, error) {
	var t models.Task
	var due sql.NullString
	var c, u string
	err := s.db.QueryRowContext(ctx, `SELECT id, title, description, status, due_date, created_at, updated_at FROM tasks WHERE id = ?`, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &due, &c, &u)
	if err != nil {
		return t, err
	}
	parseTimes(&t, due, c, u)
	return t, nil
}

func (s *Store) CreateTask(ctx context.Context, in models.Task) (models.Task, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var due *string
	if in.DueDate != nil {
		d := in.DueDate.UTC().Format(time.RFC3339)
		due = &d
	}
	res, err := s.db.ExecContext(ctx, `INSERT INTO tasks(title, description, status, due_date, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?)`,
		in.Title, in.Description, in.Status, due, now, now)
	if err != nil {
		return models.Task{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetTask(ctx, id)
}

func (s *Store) UpdateTask(ctx context.Context, id int64, in models.Task) (models.Task, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var due *string
	if in.DueDate != nil {
		d := in.DueDate.UTC().Format(time.RFC3339)
		due = &d
	}
	_, err := s.db.ExecContext(ctx, `UPDATE tasks SET title=?, description=?, status=?, due_date=?, updated_at=? WHERE id=?`,
		in.Title, in.Description, in.Status, due, now, id)
	if err != nil {
		return models.Task{}, err
	}
	return s.GetTask(ctx, id)
}

func (s *Store) DeleteTask(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	return err
}

func (s *Store) ListNotes(ctx context.Context) ([]models.Note, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, content, tags, created_at, updated_at FROM notes ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Note{}
	for rows.Next() {
		var n models.Note
		var c, u string
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Tags, &c, &u); err != nil {
			return nil, err
		}
		n.CreatedAt, _ = time.Parse(time.RFC3339, c)
		n.UpdatedAt, _ = time.Parse(time.RFC3339, u)
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Store) GetNote(ctx context.Context, id int64) (models.Note, error) {
	var n models.Note
	var c, u string
	err := s.db.QueryRowContext(ctx, `SELECT id, title, content, tags, created_at, updated_at FROM notes WHERE id = ?`, id).
		Scan(&n.ID, &n.Title, &n.Content, &n.Tags, &c, &u)
	if err != nil {
		return n, err
	}
	n.CreatedAt, _ = time.Parse(time.RFC3339, c)
	n.UpdatedAt, _ = time.Parse(time.RFC3339, u)
	return n, nil
}

func (s *Store) CreateNote(ctx context.Context, in models.Note) (models.Note, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, `INSERT INTO notes(title, content, tags, created_at, updated_at) VALUES(?, ?, ?, ?, ?)`, in.Title, in.Content, in.Tags, now, now)
	if err != nil {
		return models.Note{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetNote(ctx, id)
}

func (s *Store) UpdateNote(ctx context.Context, id int64, in models.Note) (models.Note, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `UPDATE notes SET title=?, content=?, tags=?, updated_at=? WHERE id=?`, in.Title, in.Content, in.Tags, now, id)
	if err != nil {
		return models.Note{}, err
	}
	return s.GetNote(ctx, id)
}

func (s *Store) DeleteNote(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ?`, id)
	return err
}

func parseTimes(t *models.Task, due sql.NullString, c, u string) {
	if due.Valid && strings.TrimSpace(due.String) != "" {
		d, err := time.Parse(time.RFC3339, due.String)
		if err == nil {
			t.DueDate = &d
		}
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, c)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, u)
}
