package service

import (
	"context"
	"fmt"
	"strings"

	"syncspace/backend/internal/models"
)

func (s *Service) CreateMaterial(ctx context.Context, teacherID int64, req models.Material) (models.Material, error) {
	if strings.TrimSpace(req.Title) == "" {
		return models.Material{}, fmt.Errorf("title is required")
	}
	req.TeacherID = teacherID
	return s.store.CreateMaterial(ctx, req)
}

func (s *Service) GetMaterial(ctx context.Context, id int64) (models.Material, error) {
	return s.store.GetMaterial(ctx, id)
}

func (s *Service) ListMaterialsByClassroom(ctx context.Context, classroomID int64) ([]models.Material, error) {
	return s.store.ListMaterialsByClassroom(ctx, classroomID)
}

func (s *Service) UpdateMaterial(ctx context.Context, teacherID, id int64, req models.Material) (models.Material, error) {
	m, err := s.store.GetMaterial(ctx, id)
	if err != nil {
		return models.Material{}, err
	}
	if m.TeacherID != teacherID {
		return models.Material{}, fmt.Errorf("not authorized")
	}
	return s.store.UpdateMaterial(ctx, id, req)
}

func (s *Service) DeleteMaterial(ctx context.Context, teacherID, id int64) error {
	m, err := s.store.GetMaterial(ctx, id)
	if err != nil {
		return err
	}
	if m.TeacherID != teacherID {
		return fmt.Errorf("not authorized")
	}
	return s.store.DeleteMaterial(ctx, id)
}

func (s *Service) CreateAssignment(ctx context.Context, teacherID int64, req models.Assignment) (models.Assignment, error) {
	if strings.TrimSpace(req.Title) == "" {
		return models.Assignment{}, fmt.Errorf("title is required")
	}
	req.TeacherID = teacherID
	return s.store.CreateAssignment(ctx, req)
}

func (s *Service) GetAssignment(ctx context.Context, id int64) (models.Assignment, error) {
	return s.store.GetAssignment(ctx, id)
}

func (s *Service) ListAssignmentsByClassroom(ctx context.Context, classroomID int64) ([]models.Assignment, error) {
	return s.store.ListAssignmentsByClassroom(ctx, classroomID)
}

func (s *Service) UpdateAssignment(ctx context.Context, teacherID, id int64, req models.Assignment) (models.Assignment, error) {
	a, err := s.store.GetAssignment(ctx, id)
	if err != nil {
		return models.Assignment{}, err
	}
	if a.TeacherID != teacherID {
		return models.Assignment{}, fmt.Errorf("not authorized")
	}
	return s.store.UpdateAssignment(ctx, id, req)
}

func (s *Service) DeleteAssignment(ctx context.Context, teacherID, id int64) error {
	a, err := s.store.GetAssignment(ctx, id)
	if err != nil {
		return err
	}
	if a.TeacherID != teacherID {
		return fmt.Errorf("not authorized")
	}
	return s.store.DeleteAssignment(ctx, id)
}

func (s *Service) CreateSubmission(ctx context.Context, studentID int64, assignmentID int64, content string) (models.Submission, error) {
	// Check if assignment exists
	_, err := s.store.GetAssignment(ctx, assignmentID)
	if err != nil {
		return models.Submission{}, fmt.Errorf("assignment not found")
	}

	// Check if already submitted
	existing, err := s.store.GetSubmissionByAssignmentAndStudent(ctx, assignmentID, studentID)
	if err == nil && existing.ID > 0 {
		return models.Submission{}, fmt.Errorf("already submitted")
	}

	sub := models.Submission{
		AssignmentID: assignmentID,
		StudentID:    studentID,
		Content:      content,
	}
	return s.store.CreateSubmission(ctx, sub)
}

func (s *Service) GetSubmission(ctx context.Context, id int64) (models.Submission, error) {
	return s.store.GetSubmission(ctx, id)
}

func (s *Service) ListSubmissionsByAssignment(ctx context.Context, teacherID, assignmentID int64) ([]models.Submission, error) {
	a, err := s.store.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.TeacherID != teacherID {
		return nil, fmt.Errorf("not authorized")
	}
	return s.store.ListSubmissionsByAssignment(ctx, assignmentID)
}

func (s *Service) GradeSubmission(ctx context.Context, teacherID, submissionID int64, score int, feedback string) error {
	sub, err := s.store.GetSubmission(ctx, submissionID)
	if err != nil {
		return err
	}
	a, err := s.store.GetAssignment(ctx, sub.AssignmentID)
	if err != nil {
		return err
	}
	if a.TeacherID != teacherID {
		return fmt.Errorf("not authorized")
	}
	if score < 0 || score > a.MaxScore {
		return fmt.Errorf("score must be between 0 and %d", a.MaxScore)
	}
	return s.store.GradeSubmission(ctx, submissionID, score, feedback)
}
