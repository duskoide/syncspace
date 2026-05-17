package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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

	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate(ctx context.Context) error {
	schema := `
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
