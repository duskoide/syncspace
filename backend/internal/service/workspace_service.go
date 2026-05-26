package service

import (
	"context"
	"errors"
	"strings"

	"syncspace/backend/internal/models"
	"syncspace/backend/internal/store"
)

// WorkspaceService handles business logic for workspaces
type WorkspaceService struct {
	store *store.Store
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, userID int64, req models.CreateWorkspaceRequest) (models.Workspace, error) {
	if strings.TrimSpace(req.Name) == "" {
		return models.Workspace{}, errors.New("workspace name is required")
	}

	w := models.Workspace{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		UserID:      userID,
	}

	return s.store.CreateWorkspace(ctx, w)
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, userID int64, workspaceID int64) (models.Workspace, error) {
	w, err := s.store.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return models.Workspace{}, err
	}

	// Only owner can access
	if w.UserID != userID {
		return models.Workspace{}, errors.New("access denied")
	}

	return w, nil
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID int64) ([]models.Workspace, error) {
	return s.store.ListWorkspacesByUser(ctx, userID)
}

func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, userID int64, workspaceID int64, req models.UpdateWorkspaceRequest) (models.Workspace, error) {
	// Check ownership
	w, err := s.store.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return models.Workspace{}, err
	}
	if w.UserID != userID {
		return models.Workspace{}, errors.New("access denied")
	}

	if strings.TrimSpace(req.Name) == "" {
		return models.Workspace{}, errors.New("workspace name is required")
	}

	update := models.Workspace{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}

	return s.store.UpdateWorkspace(ctx, workspaceID, update)
}

func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, userID int64, workspaceID int64) error {
	// Check ownership
	w, err := s.store.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}
	if w.UserID != userID {
		return errors.New("access denied")
	}

	return s.store.DeleteWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) IsWorkspaceOwner(ctx context.Context, userID, workspaceID int64) bool {
	w, err := s.store.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return false
	}
	return w.UserID == userID
}
