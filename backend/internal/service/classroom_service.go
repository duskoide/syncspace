package service

import (
	"context"
	"fmt"
	"strings"

	"syncspace/backend/internal/models"
)

func (s *Service) CreateClassroom(ctx context.Context, teacherID int64, req models.Classroom) (models.Classroom, error) {
	if strings.TrimSpace(req.Name) == "" {
		return models.Classroom{}, fmt.Errorf("classroom name is required")
	}
	req.TeacherID = teacherID
	return s.store.CreateClassroom(ctx, req)
}

func (s *Service) GetClassroom(ctx context.Context, id int64) (models.Classroom, error) {
	return s.store.GetClassroom(ctx, id)
}

func (s *Service) ListClassrooms(ctx context.Context, userID int64, role string) ([]models.Classroom, error) {
	if role == "teacher" {
		return s.store.ListClassrooms(ctx, userID)
	}
	// Students see classrooms they're enrolled in
	return s.store.ListStudentClassrooms(ctx, userID)
}

func (s *Service) ListAllClassrooms(ctx context.Context) ([]models.Classroom, error) {
	return s.store.ListClassrooms(ctx, 0)
}

func (s *Service) UpdateClassroom(ctx context.Context, teacherID, id int64, req models.Classroom) (models.Classroom, error) {
	c, err := s.store.GetClassroom(ctx, id)
	if err != nil {
		return models.Classroom{}, err
	}
	if c.TeacherID != teacherID {
		return models.Classroom{}, fmt.Errorf("not authorized to update this classroom")
	}
	req.TeacherID = c.TeacherID
	return s.store.UpdateClassroom(ctx, id, req)
}

func (s *Service) DeleteClassroom(ctx context.Context, teacherID, id int64) error {
	c, err := s.store.GetClassroom(ctx, id)
	if err != nil {
		return err
	}
	if c.TeacherID != teacherID {
		return fmt.Errorf("not authorized to delete this classroom")
	}
	return s.store.DeleteClassroom(ctx, id)
}

func (s *Service) RequestEnrollment(ctx context.Context, studentID, classroomID int64) (models.Enrollment, error) {
	// Check if already enrolled
	existing, err := s.store.GetEnrollmentByClassroomAndStudent(ctx, classroomID, studentID)
	if err == nil && existing.ID > 0 {
		return models.Enrollment{}, fmt.Errorf("already enrolled or pending")
	}

	e := models.Enrollment{
		ClassroomID: classroomID,
		StudentID:   studentID,
		Status:      "pending",
	}
	return s.store.CreateEnrollment(ctx, e)
}

func (s *Service) ApproveEnrollment(ctx context.Context, teacherID, enrollmentID int64) error {
	e, err := s.store.GetEnrollment(ctx, enrollmentID)
	if err != nil {
		return err
	}
	c, err := s.store.GetClassroom(ctx, e.ClassroomID)
	if err != nil {
		return err
	}
	if c.TeacherID != teacherID {
		return fmt.Errorf("not authorized to approve enrollments for this classroom")
	}
	return s.store.UpdateEnrollmentStatus(ctx, enrollmentID, "active")
}

func (s *Service) ListEnrollmentsByClassroom(ctx context.Context, teacherID, classroomID int64, status string) ([]models.Enrollment, error) {
	c, err := s.store.GetClassroom(ctx, classroomID)
	if err != nil {
		return nil, err
	}
	if c.TeacherID != teacherID {
		return nil, fmt.Errorf("not authorized")
	}
	return s.store.ListEnrollmentsByClassroom(ctx, classroomID, status)
}

func (s *Service) RemoveStudent(ctx context.Context, teacherID, classroomID, studentID int64) error {
	c, err := s.store.GetClassroom(ctx, classroomID)
	if err != nil {
		return err
	}
	if c.TeacherID != teacherID {
		return fmt.Errorf("not authorized")
	}
	return s.store.DeleteEnrollment(ctx, classroomID, studentID)
}
