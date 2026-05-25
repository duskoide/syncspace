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
	schema := `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	name TEXT NOT NULL,
	role TEXT NOT NULL CHECK(role IN ('superadmin', 'teacher', 'student')),
	status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'active', 'suspended')),
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

CREATE TABLE IF NOT EXISTS classrooms (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	teacher_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_classrooms_teacher ON classrooms(teacher_id);

CREATE TABLE IF NOT EXISTS enrollments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	classroom_id INTEGER NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,
	student_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'active', 'inactive')),
	enrolled_at TEXT NOT NULL,
	UNIQUE(classroom_id, student_id)
);
CREATE INDEX IF NOT EXISTS idx_enrollments_classroom ON enrollments(classroom_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_student ON enrollments(student_id);

CREATE TABLE IF NOT EXISTS materials (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	classroom_id INTEGER NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,
	teacher_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	content TEXT NOT NULL DEFAULT '',
	tags TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_materials_classroom ON materials(classroom_id);

CREATE TABLE IF NOT EXISTS attachments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	material_id INTEGER REFERENCES materials(id) ON DELETE CASCADE,
	submission_id INTEGER REFERENCES submissions(id) ON DELETE CASCADE,
	filename TEXT NOT NULL,
	original_name TEXT NOT NULL,
	mime_type TEXT NOT NULL,
	file_size INTEGER NOT NULL,
	file_path TEXT NOT NULL,
	uploaded_by INTEGER NOT NULL REFERENCES users(id),
	created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_attachments_material ON attachments(material_id);
CREATE INDEX IF NOT EXISTS idx_attachments_submission ON attachments(submission_id);

CREATE TABLE IF NOT EXISTS assignments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	classroom_id INTEGER NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,
	teacher_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	due_date TEXT NOT NULL,
	max_score INTEGER DEFAULT 100,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_assignments_classroom ON assignments(classroom_id);

CREATE TABLE IF NOT EXISTS submissions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	assignment_id INTEGER NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
	student_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	content TEXT NOT NULL DEFAULT '',
	score INTEGER,
	feedback TEXT DEFAULT '',
	submitted_at TEXT NOT NULL,
	graded_at TEXT,
	UNIQUE(assignment_id, student_id)
);
CREATE INDEX IF NOT EXISTS idx_submissions_assignment ON submissions(assignment_id);
CREATE INDEX IF NOT EXISTS idx_submissions_student ON submissions(student_id);

CREATE TABLE IF NOT EXISTS collaborative_notes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	material_id INTEGER REFERENCES materials(id) ON DELETE SET NULL,
	classroom_id INTEGER NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,
	created_by INTEGER NOT NULL REFERENCES users(id),
	title TEXT NOT NULL,
	content TEXT NOT NULL DEFAULT '',
	is_public BOOLEAN DEFAULT 1,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_notes_classroom ON collaborative_notes(classroom_id);

CREATE TABLE IF NOT EXISTS note_contributors (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	note_id INTEGER NOT NULL REFERENCES collaborative_notes(id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	contributed_at TEXT NOT NULL,
	UNIQUE(note_id, user_id)
);

CREATE TABLE IF NOT EXISTS discussions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	classroom_id INTEGER NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,
	material_id INTEGER REFERENCES materials(id) ON DELETE CASCADE,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	message TEXT NOT NULL,
	parent_id INTEGER REFERENCES discussions(id) ON DELETE CASCADE,
	created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_discussions_classroom ON discussions(classroom_id);

-- Keep old tables for backward compatibility
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

// Classroom methods

func (s *Store) CreateClassroom(ctx context.Context, c models.Classroom) (models.Classroom, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO classrooms(name, description, teacher_id, created_at, updated_at) VALUES(?, ?, ?, ?, ?)`,
		c.Name, c.Description, c.TeacherID, now, now)
	if err != nil {
		return models.Classroom{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetClassroom(ctx, id)
}

func (s *Store) GetClassroom(ctx context.Context, id int64) (models.Classroom, error) {
	var c models.Classroom
	var cr, up string
	err := s.db.QueryRowContext(ctx,
		`SELECT c.id, c.name, c.description, c.teacher_id, u.name, c.created_at, c.updated_at 
		 FROM classrooms c JOIN users u ON c.teacher_id = u.id WHERE c.id = ?`, id).
		Scan(&c.ID, &c.Name, &c.Description, &c.TeacherID, &c.TeacherName, &cr, &up)
	if err != nil {
		return c, err
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339, cr)
	c.UpdatedAt, _ = time.Parse(time.RFC3339, up)
	return c, nil
}

func (s *Store) ListClassrooms(ctx context.Context, teacherID int64) ([]models.Classroom, error) {
	query := `SELECT c.id, c.name, c.description, c.teacher_id, u.name, c.created_at, c.updated_at 
		      FROM classrooms c JOIN users u ON c.teacher_id = u.id`
	args := []interface{}{}
	if teacherID > 0 {
		query += ` WHERE c.teacher_id = ?`
		args = append(args, teacherID)
	}
	query += ` ORDER BY c.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Classroom{}
	for rows.Next() {
		var c models.Classroom
		var cr, up string
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.TeacherID, &c.TeacherName, &cr, &up); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, cr)
		c.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) ListStudentClassrooms(ctx context.Context, studentID int64) ([]models.Classroom, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.name, c.description, c.teacher_id, u.name, c.created_at, c.updated_at 
		 FROM classrooms c 
		 JOIN users u ON c.teacher_id = u.id 
		 JOIN enrollments e ON c.id = e.classroom_id 
		 WHERE e.student_id = ? AND e.status = 'active'
		 ORDER BY c.created_at DESC`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Classroom{}
	for rows.Next() {
		var c models.Classroom
		var cr, up string
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.TeacherID, &c.TeacherName, &cr, &up); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, cr)
		c.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) UpdateClassroom(ctx context.Context, id int64, c models.Classroom) (models.Classroom, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`UPDATE classrooms SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		c.Name, c.Description, now, id)
	if err != nil {
		return models.Classroom{}, err
	}
	return s.GetClassroom(ctx, id)
}

func (s *Store) DeleteClassroom(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM classrooms WHERE id = ?`, id)
	return err
}

// Enrollment methods

func (s *Store) CreateEnrollment(ctx context.Context, e models.Enrollment) (models.Enrollment, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO enrollments(classroom_id, student_id, status, enrolled_at) VALUES(?, ?, ?, ?)`,
		e.ClassroomID, e.StudentID, e.Status, now)
	if err != nil {
		return models.Enrollment{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetEnrollment(ctx, id)
}

func (s *Store) GetEnrollment(ctx context.Context, id int64) (models.Enrollment, error) {
	var e models.Enrollment
	var en string
	err := s.db.QueryRowContext(ctx,
		`SELECT e.id, e.classroom_id, e.student_id, u.name, u.email, e.status, e.enrolled_at
		 FROM enrollments e JOIN users u ON e.student_id = u.id WHERE e.id = ?`, id).
		Scan(&e.ID, &e.ClassroomID, &e.StudentID, &e.StudentName, &e.StudentEmail, &e.Status, &en)
	if err != nil {
		return e, err
	}
	e.EnrolledAt, _ = time.Parse(time.RFC3339, en)
	return e, nil
}

func (s *Store) GetEnrollmentByClassroomAndStudent(ctx context.Context, classroomID, studentID int64) (models.Enrollment, error) {
	var e models.Enrollment
	var en string
	err := s.db.QueryRowContext(ctx,
		`SELECT e.id, e.classroom_id, e.student_id, u.name, u.email, e.status, e.enrolled_at
		 FROM enrollments e JOIN users u ON e.student_id = u.id 
		 WHERE e.classroom_id = ? AND e.student_id = ?`, classroomID, studentID).
		Scan(&e.ID, &e.ClassroomID, &e.StudentID, &e.StudentName, &e.StudentEmail, &e.Status, &en)
	if err != nil {
		return e, err
	}
	e.EnrolledAt, _ = time.Parse(time.RFC3339, en)
	return e, nil
}

func (s *Store) ListEnrollmentsByClassroom(ctx context.Context, classroomID int64, status string) ([]models.Enrollment, error) {
	query := `SELECT e.id, e.classroom_id, e.student_id, u.name, u.email, e.status, e.enrolled_at
		      FROM enrollments e JOIN users u ON e.student_id = u.id WHERE e.classroom_id = ?`
	args := []interface{}{classroomID}
	if status != "" {
		query += ` AND e.status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY e.enrolled_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Enrollment{}
	for rows.Next() {
		var e models.Enrollment
		var en string
		if err := rows.Scan(&e.ID, &e.ClassroomID, &e.StudentID, &e.StudentName, &e.StudentEmail, &e.Status, &en); err != nil {
			return nil, err
		}
		e.EnrolledAt, _ = time.Parse(time.RFC3339, en)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) UpdateEnrollmentStatus(ctx context.Context, id int64, status string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE enrollments SET status = ? WHERE id = ?`,
		status, id)
	return err
}

func (s *Store) DeleteEnrollment(ctx context.Context, classroomID, studentID int64) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM enrollments WHERE classroom_id = ? AND student_id = ?`,
		classroomID, studentID)
	return err
}

// Material methods

func (s *Store) CreateMaterial(ctx context.Context, m models.Material) (models.Material, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO materials(classroom_id, teacher_id, title, content, tags, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?)`,
		m.ClassroomID, m.TeacherID, m.Title, m.Content, m.Tags, now, now)
	if err != nil {
		return models.Material{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetMaterial(ctx, id)
}

func (s *Store) GetMaterial(ctx context.Context, id int64) (models.Material, error) {
	var m models.Material
	var c, up string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, classroom_id, teacher_id, title, content, tags, created_at, updated_at FROM materials WHERE id = ?`, id).
		Scan(&m.ID, &m.ClassroomID, &m.TeacherID, &m.Title, &m.Content, &m.Tags, &c, &up)
	if err != nil {
		return m, err
	}
	m.CreatedAt, _ = time.Parse(time.RFC3339, c)
	m.UpdatedAt, _ = time.Parse(time.RFC3339, up)
	return m, nil
}

func (s *Store) ListMaterialsByClassroom(ctx context.Context, classroomID int64) ([]models.Material, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, classroom_id, teacher_id, title, content, tags, created_at, updated_at 
		 FROM materials WHERE classroom_id = ? ORDER BY created_at DESC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Material{}
	for rows.Next() {
		var m models.Material
		var c, up string
		if err := rows.Scan(&m.ID, &m.ClassroomID, &m.TeacherID, &m.Title, &m.Content, &m.Tags, &c, &up); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, c)
		m.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) UpdateMaterial(ctx context.Context, id int64, m models.Material) (models.Material, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`UPDATE materials SET title = ?, content = ?, tags = ?, updated_at = ? WHERE id = ?`,
		m.Title, m.Content, m.Tags, now, id)
	if err != nil {
		return models.Material{}, err
	}
	return s.GetMaterial(ctx, id)
}

func (s *Store) DeleteMaterial(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM materials WHERE id = ?`, id)
	return err
}

// Assignment methods

func (s *Store) CreateAssignment(ctx context.Context, a models.Assignment) (models.Assignment, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	due := a.DueDate.UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO assignments(classroom_id, teacher_id, title, description, due_date, max_score, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ClassroomID, a.TeacherID, a.Title, a.Description, due, a.MaxScore, now, now)
	if err != nil {
		return models.Assignment{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetAssignment(ctx, id)
}

func (s *Store) GetAssignment(ctx context.Context, id int64) (models.Assignment, error) {
	var a models.Assignment
	var due, c, up string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, classroom_id, teacher_id, title, description, due_date, max_score, created_at, updated_at FROM assignments WHERE id = ?`, id).
		Scan(&a.ID, &a.ClassroomID, &a.TeacherID, &a.Title, &a.Description, &due, &a.MaxScore, &c, &up)
	if err != nil {
		return a, err
	}
	a.DueDate, _ = time.Parse(time.RFC3339, due)
	a.CreatedAt, _ = time.Parse(time.RFC3339, c)
	a.UpdatedAt, _ = time.Parse(time.RFC3339, up)
	return a, nil
}

func (s *Store) ListAssignmentsByClassroom(ctx context.Context, classroomID int64) ([]models.Assignment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, classroom_id, teacher_id, title, description, due_date, max_score, created_at, updated_at 
		 FROM assignments WHERE classroom_id = ? ORDER BY due_date ASC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Assignment{}
	for rows.Next() {
		var a models.Assignment
		var due, c, up string
		if err := rows.Scan(&a.ID, &a.ClassroomID, &a.TeacherID, &a.Title, &a.Description, &due, &a.MaxScore, &c, &up); err != nil {
			return nil, err
		}
		a.DueDate, _ = time.Parse(time.RFC3339, due)
		a.CreatedAt, _ = time.Parse(time.RFC3339, c)
		a.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) UpdateAssignment(ctx context.Context, id int64, a models.Assignment) (models.Assignment, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	due := a.DueDate.UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`UPDATE assignments SET title = ?, description = ?, due_date = ?, max_score = ?, updated_at = ? WHERE id = ?`,
		a.Title, a.Description, due, a.MaxScore, now, id)
	if err != nil {
		return models.Assignment{}, err
	}
	return s.GetAssignment(ctx, id)
}

func (s *Store) DeleteAssignment(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM assignments WHERE id = ?`, id)
	return err
}

// Submission methods

func (s *Store) CreateSubmission(ctx context.Context, sub models.Submission) (models.Submission, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO submissions(assignment_id, student_id, content, score, feedback, submitted_at, graded_at) VALUES(?, ?, ?, ?, ?, ?, ?)`,
		sub.AssignmentID, sub.StudentID, sub.Content, sub.Score, sub.Feedback, now, nil)
	if err != nil {
		return models.Submission{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetSubmission(ctx, id)
}

func (s *Store) GetSubmission(ctx context.Context, id int64) (models.Submission, error) {
	var sub models.Submission
	var graded sql.NullString
	var c string
	err := s.db.QueryRowContext(ctx,
		`SELECT s.id, s.assignment_id, s.student_id, u.name, s.content, s.score, s.feedback, s.submitted_at, s.graded_at
		 FROM submissions s JOIN users u ON s.student_id = u.id WHERE s.id = ?`, id).
		Scan(&sub.ID, &sub.AssignmentID, &sub.StudentID, &sub.StudentName, &sub.Content, &sub.Score, &sub.Feedback, &c, &graded)
	if err != nil {
		return sub, err
	}
	sub.SubmittedAt, _ = time.Parse(time.RFC3339, c)
	if graded.Valid {
		g, _ := time.Parse(time.RFC3339, graded.String)
		sub.GradedAt = &g
	}
	return sub, nil
}

func (s *Store) ListSubmissionsByAssignment(ctx context.Context, assignmentID int64) ([]models.Submission, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT s.id, s.assignment_id, s.student_id, u.name, s.content, s.score, s.feedback, s.submitted_at, s.graded_at
		 FROM submissions s JOIN users u ON s.student_id = u.id WHERE s.assignment_id = ? ORDER BY s.submitted_at DESC`, assignmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Submission{}
	for rows.Next() {
		var sub models.Submission
		var graded sql.NullString
		var c string
		if err := rows.Scan(&sub.ID, &sub.AssignmentID, &sub.StudentID, &sub.StudentName, &sub.Content, &sub.Score, &sub.Feedback, &c, &graded); err != nil {
			return nil, err
		}
		sub.SubmittedAt, _ = time.Parse(time.RFC3339, c)
		if graded.Valid {
			g, _ := time.Parse(time.RFC3339, graded.String)
			sub.GradedAt = &g
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (s *Store) GetSubmissionByAssignmentAndStudent(ctx context.Context, assignmentID, studentID int64) (models.Submission, error) {
	var sub models.Submission
	var graded sql.NullString
	var c string
	err := s.db.QueryRowContext(ctx,
		`SELECT s.id, s.assignment_id, s.student_id, u.name, s.content, s.score, s.feedback, s.submitted_at, s.graded_at
		 FROM submissions s JOIN users u ON s.student_id = u.id WHERE s.assignment_id = ? AND s.student_id = ?`, assignmentID, studentID).
		Scan(&sub.ID, &sub.AssignmentID, &sub.StudentID, &sub.StudentName, &sub.Content, &sub.Score, &sub.Feedback, &c, &graded)
	if err != nil {
		return sub, err
	}
	sub.SubmittedAt, _ = time.Parse(time.RFC3339, c)
	if graded.Valid {
		g, _ := time.Parse(time.RFC3339, graded.String)
		sub.GradedAt = &g
	}
	return sub, nil
}

func (s *Store) GradeSubmission(ctx context.Context, id int64, score int, feedback string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`UPDATE submissions SET score = ?, feedback = ?, graded_at = ? WHERE id = ?`,
		score, feedback, now, id)
	return err
}

// Attachment methods

func (s *Store) CreateAttachment(ctx context.Context, a models.Attachment) (models.Attachment, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO attachments(material_id, submission_id, filename, original_name, mime_type, file_size, file_path, uploaded_by, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.MaterialID, a.SubmissionID, a.Filename, a.OriginalName, a.MimeType, a.FileSize, a.FilePath, a.UploadedBy, now)
	if err != nil {
		return models.Attachment{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetAttachment(ctx, id)
}

func (s *Store) GetAttachment(ctx context.Context, id int64) (models.Attachment, error) {
	var a models.Attachment
	var c string
	var mid, sid sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT id, material_id, submission_id, filename, original_name, mime_type, file_size, file_path, uploaded_by, created_at FROM attachments WHERE id = ?`, id).
		Scan(&a.ID, &mid, &sid, &a.Filename, &a.OriginalName, &a.MimeType, &a.FileSize, &a.FilePath, &a.UploadedBy, &c)
	if err != nil {
		return a, err
	}
	if mid.Valid {
		a.MaterialID = &mid.Int64
	}
	if sid.Valid {
		a.SubmissionID = &sid.Int64
	}
	a.CreatedAt, _ = time.Parse(time.RFC3339, c)
	return a, nil
}

func (s *Store) ListAttachmentsByMaterial(ctx context.Context, materialID int64) ([]models.Attachment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, material_id, submission_id, filename, original_name, mime_type, file_size, file_path, uploaded_by, created_at 
		 FROM attachments WHERE material_id = ? ORDER BY created_at DESC`, materialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Attachment{}
	for rows.Next() {
		var a models.Attachment
		var c string
		var mid, sid sql.NullInt64
		if err := rows.Scan(&a.ID, &mid, &sid, &a.Filename, &a.OriginalName, &a.MimeType, &a.FileSize, &a.FilePath, &a.UploadedBy, &c); err != nil {
			return nil, err
		}
		if mid.Valid {
			a.MaterialID = &mid.Int64
		}
		if sid.Valid {
			a.SubmissionID = &sid.Int64
		}
		a.CreatedAt, _ = time.Parse(time.RFC3339, c)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) ListAttachmentsBySubmission(ctx context.Context, submissionID int64) ([]models.Attachment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, material_id, submission_id, filename, original_name, mime_type, file_size, file_path, uploaded_by, created_at 
		 FROM attachments WHERE submission_id = ? ORDER BY created_at DESC`, submissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Attachment{}
	for rows.Next() {
		var a models.Attachment
		var c string
		var mid, sid sql.NullInt64
		if err := rows.Scan(&a.ID, &mid, &sid, &a.Filename, &a.OriginalName, &a.MimeType, &a.FileSize, &a.FilePath, &a.UploadedBy, &c); err != nil {
			return nil, err
		}
		if mid.Valid {
			a.MaterialID = &mid.Int64
		}
		if sid.Valid {
			a.SubmissionID = &sid.Int64
		}
		a.CreatedAt, _ = time.Parse(time.RFC3339, c)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) DeleteAttachment(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM attachments WHERE id = ?`, id)
	return err
}

// Collaborative Note methods

func (s *Store) CreateCollaborativeNote(ctx context.Context, n models.CollaborativeNote) (models.CollaborativeNote, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO collaborative_notes(material_id, classroom_id, created_by, title, content, is_public, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		n.MaterialID, n.ClassroomID, n.CreatedBy, n.Title, n.Content, n.IsPublic, now, now)
	if err != nil {
		return models.CollaborativeNote{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetCollaborativeNote(ctx, id)
}

func (s *Store) GetCollaborativeNote(ctx context.Context, id int64) (models.CollaborativeNote, error) {
	var n models.CollaborativeNote
	var mid sql.NullInt64
	var c, up string
	err := s.db.QueryRowContext(ctx,
		`SELECT n.id, n.material_id, n.classroom_id, n.created_by, u.name, n.title, n.content, n.is_public, n.created_at, n.updated_at
		 FROM collaborative_notes n JOIN users u ON n.created_by = u.id WHERE n.id = ?`, id).
		Scan(&n.ID, &mid, &n.ClassroomID, &n.CreatedBy, &n.CreatorName, &n.Title, &n.Content, &n.IsPublic, &c, &up)
	if err != nil {
		return n, err
	}
	if mid.Valid {
		n.MaterialID = &mid.Int64
	}
	n.CreatedAt, _ = time.Parse(time.RFC3339, c)
	n.UpdatedAt, _ = time.Parse(time.RFC3339, up)
	return n, nil
}

func (s *Store) ListCollaborativeNotesByClassroom(ctx context.Context, classroomID int64) ([]models.CollaborativeNote, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT n.id, n.material_id, n.classroom_id, n.created_by, u.name, n.title, n.content, n.is_public, n.created_at, n.updated_at
		 FROM collaborative_notes n JOIN users u ON n.created_by = u.id WHERE n.classroom_id = ? ORDER BY n.updated_at DESC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.CollaborativeNote{}
	for rows.Next() {
		var n models.CollaborativeNote
		var mid sql.NullInt64
		var c, up string
		if err := rows.Scan(&n.ID, &mid, &n.ClassroomID, &n.CreatedBy, &n.CreatorName, &n.Title, &n.Content, &n.IsPublic, &c, &up); err != nil {
			return nil, err
		}
		if mid.Valid {
			n.MaterialID = &mid.Int64
		}
		n.CreatedAt, _ = time.Parse(time.RFC3339, c)
		n.UpdatedAt, _ = time.Parse(time.RFC3339, up)
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Store) UpdateCollaborativeNote(ctx context.Context, id int64, n models.CollaborativeNote) (models.CollaborativeNote, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`UPDATE collaborative_notes SET title = ?, content = ?, is_public = ?, updated_at = ? WHERE id = ?`,
		n.Title, n.Content, n.IsPublic, now, id)
	if err != nil {
		return models.CollaborativeNote{}, err
	}
	return s.GetCollaborativeNote(ctx, id)
}

func (s *Store) DeleteCollaborativeNote(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM collaborative_notes WHERE id = ?`, id)
	return err
}

func (s *Store) AddNoteContributor(ctx context.Context, noteID, userID int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO note_contributors(note_id, user_id, contributed_at) VALUES(?, ?, ?)`,
		noteID, userID, now)
	return err
}

func (s *Store) GetNoteContributors(ctx context.Context, noteID int64) ([]models.User, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT u.id, u.email, u.name, u.role, u.status, u.created_at, u.updated_at
		 FROM note_contributors nc JOIN users u ON nc.user_id = u.id WHERE nc.note_id = ?`, noteID)
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

// Discussion methods

func (s *Store) CreateDiscussion(ctx context.Context, d models.Discussion) (models.Discussion, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO discussions(classroom_id, material_id, user_id, message, parent_id, created_at) VALUES(?, ?, ?, ?, ?, ?)`,
		d.ClassroomID, d.MaterialID, d.UserID, d.Message, d.ParentID, now)
	if err != nil {
		return models.Discussion{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetDiscussion(ctx, id)
}

func (s *Store) GetDiscussion(ctx context.Context, id int64) (models.Discussion, error) {
	var d models.Discussion
	var mid, pid sql.NullInt64
	var c string
	err := s.db.QueryRowContext(ctx,
		`SELECT d.id, d.classroom_id, d.material_id, d.user_id, u.name, d.message, d.parent_id, d.created_at
		 FROM discussions d JOIN users u ON d.user_id = u.id WHERE d.id = ?`, id).
		Scan(&d.ID, &d.ClassroomID, &mid, &d.UserID, &d.UserName, &d.Message, &pid, &c)
	if err != nil {
		return d, err
	}
	if mid.Valid {
		d.MaterialID = &mid.Int64
	}
	if pid.Valid {
		d.ParentID = &pid.Int64
	}
	d.CreatedAt, _ = time.Parse(time.RFC3339, c)
	return d, nil
}

func (s *Store) ListDiscussionsByClassroom(ctx context.Context, classroomID int64, limit, offset int) ([]models.Discussion, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT d.id, d.classroom_id, d.material_id, d.user_id, u.name, d.message, d.parent_id, d.created_at
		 FROM discussions d JOIN users u ON d.user_id = u.id WHERE d.classroom_id = ? AND d.parent_id IS NULL
		 ORDER BY d.created_at DESC LIMIT ? OFFSET ?`, classroomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Discussion{}
	for rows.Next() {
		var d models.Discussion
		var mid, pid sql.NullInt64
		var c string
		if err := rows.Scan(&d.ID, &d.ClassroomID, &mid, &d.UserID, &d.UserName, &d.Message, &pid, &c); err != nil {
			return nil, err
		}
		if mid.Valid {
			d.MaterialID = &mid.Int64
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
		`SELECT d.id, d.classroom_id, d.material_id, d.user_id, u.name, d.message, d.parent_id, d.created_at
		 FROM discussions d JOIN users u ON d.user_id = u.id WHERE d.parent_id = ? ORDER BY d.created_at ASC`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Discussion{}
	for rows.Next() {
		var d models.Discussion
		var mid, pid sql.NullInt64
		var c string
		if err := rows.Scan(&d.ID, &d.ClassroomID, &mid, &d.UserID, &d.UserName, &d.Message, &pid, &c); err != nil {
			return nil, err
		}
		if mid.Valid {
			d.MaterialID = &mid.Int64
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
