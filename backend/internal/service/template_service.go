package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"syncspace/backend/internal/models"
	"syncspace/backend/internal/store"
)

// TemplateService handles business logic for templates
type TemplateService struct {
	store *store.Store
}

// CreateTemplate creates a template from a workspace
func (s *TemplateService) CreateTemplate(ctx context.Context, creatorID int64, req models.CreateTemplateRequest) (models.Template, error) {
	if req.Type != "workspace" {
		return models.Template{}, errors.New("type must be 'workspace'")
	}
	if strings.TrimSpace(req.Name) == "" {
		return models.Template{}, errors.New("template name is required")
	}
	if req.Visibility != "public" && req.Visibility != "link" {
		req.Visibility = "public"
	}

	// Verify ownership and capture snapshot
	w, err := s.store.GetWorkspace(ctx, req.SourceID)
	if err != nil {
		return models.Template{}, errors.New("workspace not found")
	}
	if w.UserID != creatorID {
		return models.Template{}, errors.New("access denied")
	}

	// Get all notes in workspace
	notes, err := s.store.ListNotesByWorkspace(ctx, req.SourceID)
	if err != nil {
		return models.Template{}, err
	}

	snapshot := models.TemplateSnapshot{
		WorkspaceID: w.ID,
		Name:        w.Name,
		Description: w.Description,
		Notes:       notes,
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return models.Template{}, err
	}

	t := models.Template{
		Type:            req.Type,
		SourceID:        req.SourceID,
		CreatorID:       creatorID,
		Name:            strings.TrimSpace(req.Name),
		Description:     strings.TrimSpace(req.Description),
		Visibility:      req.Visibility,
		ContentSnapshot: string(snapshotJSON),
		IsHidden:        false,
	}

	return s.store.CreateTemplate(ctx, t)
}

// UpdateTemplate updates template metadata (name, description, visibility)
func (s *TemplateService) UpdateTemplate(ctx context.Context, creatorID int64, templateID int64, req models.UpdateTemplateRequest) (models.Template, error) {
	t, err := s.store.GetTemplate(ctx, templateID)
	if err != nil {
		return models.Template{}, err
	}

	// Only creator or superadmin can update
	if t.CreatorID != creatorID {
		// Check if user is superadmin
		u, err := s.store.GetUserByID(ctx, creatorID)
		if err != nil || u.Role != "superadmin" {
			return models.Template{}, errors.New("access denied")
		}
	}

	if strings.TrimSpace(req.Name) == "" {
		return models.Template{}, errors.New("template name is required")
	}
	if req.Visibility != "" && req.Visibility != "public" && req.Visibility != "link" {
		return models.Template{}, errors.New("visibility must be 'public' or 'link'")
	}

	t.Name = strings.TrimSpace(req.Name)
	t.Description = strings.TrimSpace(req.Description)
	if req.Visibility != "" {
		t.Visibility = req.Visibility
	}

	return s.store.UpdateTemplate(ctx, templateID, t)
}

// UpdateTemplateContent re-snapshots the current state of the source workspace
func (s *TemplateService) UpdateTemplateContent(ctx context.Context, creatorID int64, templateID int64) (models.Template, error) {
	t, err := s.store.GetTemplate(ctx, templateID)
	if err != nil {
		return models.Template{}, err
	}

	if t.CreatorID != creatorID {
		return models.Template{}, errors.New("access denied")
	}

	// Capture new snapshot
	w, err := s.store.GetWorkspace(ctx, t.SourceID)
	if err != nil {
		return models.Template{}, errors.New("source workspace not found")
	}
	if w.UserID != creatorID {
		return models.Template{}, errors.New("access denied")
	}

	notes, err := s.store.ListNotesByWorkspace(ctx, t.SourceID)
	if err != nil {
		return models.Template{}, err
	}

	snapshot := models.TemplateSnapshot{
		WorkspaceID: w.ID,
		Name:        w.Name,
		Description: w.Description,
		Notes:       notes,
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return models.Template{}, err
	}

	t.ContentSnapshot = string(snapshotJSON)
	return s.store.UpdateTemplate(ctx, templateID, t)
}

// GetTemplate retrieves a template by ID
func (s *TemplateService) GetTemplate(ctx context.Context, userID int64, templateID int64) (models.Template, error) {
	t, err := s.store.GetTemplate(ctx, templateID)
	if err != nil {
		return models.Template{}, err
	}

	// Check access
	if t.IsHidden {
		// Only creator or superadmin can see hidden templates
		u, _ := s.store.GetUserByID(ctx, userID)
		if t.CreatorID != userID && u.Role != "superadmin" {
			return models.Template{}, errors.New("template not found")
		}
	}

	// Link-only templates can be accessed by anyone with the ID
	// Public templates are accessible to all
	return t, nil
}

// ListTemplates lists public templates with optional search
func (s *TemplateService) ListTemplates(ctx context.Context, search string) ([]models.Template, error) {
	return s.store.ListTemplates(ctx, "public", search, true)
}

// ListMyTemplates lists templates created by the user
func (s *TemplateService) ListMyTemplates(ctx context.Context, creatorID int64) ([]models.Template, error) {
	return s.store.ListTemplatesByCreator(ctx, creatorID)
}

// ListAllTemplatesForAdmin lists all templates for superadmin moderation
func (s *TemplateService) ListAllTemplatesForAdmin(ctx context.Context) ([]models.Template, error) {
	return s.store.ListTemplates(ctx, "", "", false)
}

// DeleteTemplate deletes a template
func (s *TemplateService) DeleteTemplate(ctx context.Context, userID int64, templateID int64) error {
	t, err := s.store.GetTemplate(ctx, templateID)
	if err != nil {
		return err
	}

	// Only creator or superadmin can delete
	if t.CreatorID != userID {
		u, err := s.store.GetUserByID(ctx, userID)
		if err != nil || u.Role != "superadmin" {
			return errors.New("access denied")
		}
	}

	return s.store.DeleteTemplate(ctx, templateID)
}

// SetTemplateHidden allows superadmin to hide/unhide templates
func (s *TemplateService) SetTemplateHidden(ctx context.Context, adminID int64, templateID int64, isHidden bool) error {
	// Verify admin
	u, err := s.store.GetUserByID(ctx, adminID)
	if err != nil || u.Role != "superadmin" {
		return errors.New("access denied")
	}

	return s.store.UpdateTemplateHidden(ctx, templateID, isHidden)
}

// CloneTemplate clones a workspace template into the user's account
func (s *TemplateService) CloneTemplate(ctx context.Context, userID int64, templateID int64, req models.CloneTemplateRequest) (*models.Workspace, *models.Note, error) {
	t, err := s.store.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, nil, err
	}

	// Check access
	if t.IsHidden {
		return nil, nil, errors.New("template not found")
	}

	// Parse snapshot
	var snapshot models.TemplateSnapshot
	if err := json.Unmarshal([]byte(t.ContentSnapshot), &snapshot); err != nil {
		return nil, nil, err
	}

	// Create new workspace
	w, err := s.store.CreateWorkspace(ctx, models.Workspace{
		Name:        snapshot.Name + " (Copy)",
		Description: snapshot.Description,
		UserID:      userID,
	})
	if err != nil {
		return nil, nil, err
	}

	// Clone all notes
	for _, note := range snapshot.Notes {
		_, err := s.store.CreateNote(ctx, models.Note{
			WorkspaceID: w.ID,
			Title:       note.Title,
			Content:     note.Content,
			CreatedBy:   userID,
		})
		if err != nil {
			// Log error but continue
			continue
		}
	}

	return &w, nil, nil
}
