package service

import (
	"context"
	"fmt"
	"strings"

	"syncspace/backend/internal/models"
)

func (s *Service) CreateCollaborativeNote(ctx context.Context, userID int64, req models.CollaborativeNote) (models.CollaborativeNote, error) {
	if strings.TrimSpace(req.Title) == "" {
		return models.CollaborativeNote{}, fmt.Errorf("title is required")
	}
	req.CreatedBy = userID
	note, err := s.store.CreateCollaborativeNote(ctx, req)
	if err != nil {
		return models.CollaborativeNote{}, err
	}
	// Add creator as first contributor
	_ = s.store.AddNoteContributor(ctx, note.ID, userID)
	return note, nil
}

func (s *Service) GetCollaborativeNote(ctx context.Context, id int64) (models.CollaborativeNote, error) {
	return s.store.GetCollaborativeNote(ctx, id)
}

func (s *Service) ListCollaborativeNotesByClassroom(ctx context.Context, classroomID int64) ([]models.CollaborativeNote, error) {
	return s.store.ListCollaborativeNotesByClassroom(ctx, classroomID)
}

func (s *Service) UpdateCollaborativeNote(ctx context.Context, userID, id int64, req models.CollaborativeNote) (models.CollaborativeNote, error) {
	note, err := s.store.GetCollaborativeNote(ctx, id)
	if err != nil {
		return models.CollaborativeNote{}, err
	}
	// Only creator or contributors can update
	if note.CreatedBy != userID {
		// Check if contributor
		contributors, err := s.store.GetNoteContributors(ctx, id)
		if err != nil {
			return models.CollaborativeNote{}, err
		}
		isContributor := false
		for _, c := range contributors {
			if c.ID == userID {
				isContributor = true
				break
			}
		}
		if !isContributor {
			return models.CollaborativeNote{}, fmt.Errorf("not authorized")
		}
	}
	_ = s.store.AddNoteContributor(ctx, id, userID)
	return s.store.UpdateCollaborativeNote(ctx, id, req)
}

func (s *Service) DeleteCollaborativeNote(ctx context.Context, userID, id int64) error {
	note, err := s.store.GetCollaborativeNote(ctx, id)
	if err != nil {
		return err
	}
	if note.CreatedBy != userID {
		return fmt.Errorf("not authorized")
	}
	return s.store.DeleteCollaborativeNote(ctx, id)
}

func (s *Service) GetNoteContributors(ctx context.Context, noteID int64) ([]models.User, error) {
	return s.store.GetNoteContributors(ctx, noteID)
}

func (s *Service) CreateDiscussion(ctx context.Context, userID int64, req models.Discussion) (models.Discussion, error) {
	if strings.TrimSpace(req.Message) == "" {
		return models.Discussion{}, fmt.Errorf("message is required")
	}
	req.UserID = userID
	return s.store.CreateDiscussion(ctx, req)
}

func (s *Service) GetDiscussion(ctx context.Context, id int64) (models.Discussion, error) {
	return s.store.GetDiscussion(ctx, id)
}

func (s *Service) ListDiscussionsByClassroom(ctx context.Context, classroomID int64, limit, offset int) ([]models.Discussion, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.store.ListDiscussionsByClassroom(ctx, classroomID, limit, offset)
}

func (s *Service) ListDiscussionReplies(ctx context.Context, parentID int64) ([]models.Discussion, error) {
	return s.store.ListDiscussionReplies(ctx, parentID)
}
