package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"syncspace/backend/internal/models"
	"syncspace/backend/internal/store"
)

// NoteService handles business logic for notes
type NoteService struct {
	store *store.Store
}

func (s *NoteService) CreateNote(ctx context.Context, userID int64, req models.CreateNoteRequest) (models.Note, error) {
	if req.WorkspaceID == 0 {
		return models.Note{}, errors.New("workspace_id is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return models.Note{}, errors.New("note title is required")
	}

	// Verify workspace ownership
	w, err := s.store.GetWorkspace(ctx, req.WorkspaceID)
	if err != nil {
		return models.Note{}, errors.New("workspace not found")
	}
	if w.UserID != userID {
		return models.Note{}, errors.New("access denied")
	}

	n := models.Note{
		WorkspaceID: req.WorkspaceID,
		Title:       strings.TrimSpace(req.Title),
		Content:     "", // Empty content initially
		CreatedBy:   userID,
	}

	return s.store.CreateNote(ctx, n)
}

func (s *NoteService) GetNote(ctx context.Context, userID int64, noteID int64) (models.Note, error) {
	n, err := s.store.GetNote(ctx, noteID)
	if err != nil {
		return models.Note{}, err
	}

	// Verify workspace ownership
	if !s.isWorkspaceOwner(ctx, userID, n.WorkspaceID) {
		return models.Note{}, errors.New("access denied")
	}

	return n, nil
}

func (s *NoteService) ListNotesByWorkspace(ctx context.Context, userID int64, workspaceID int64) ([]models.Note, error) {
	// Verify workspace ownership
	if !s.isWorkspaceOwner(ctx, userID, workspaceID) {
		return nil, errors.New("access denied")
	}

	return s.store.ListNotesByWorkspace(ctx, workspaceID)
}

func (s *NoteService) UpdateNote(ctx context.Context, userID int64, noteID int64, req models.UpdateNoteRequest) (models.Note, error) {
	// Get note and verify access
	n, err := s.store.GetNote(ctx, noteID)
	if err != nil {
		return models.Note{}, err
	}

	if !s.isWorkspaceOwner(ctx, userID, n.WorkspaceID) {
		return models.Note{}, errors.New("access denied")
	}

	if strings.TrimSpace(req.Title) == "" {
		return models.Note{}, errors.New("note title is required")
	}

	update := models.Note{
		Title:   strings.TrimSpace(req.Title),
		Content: sanitizeHTML(req.Content), // Sanitize HTML from TipTap
	}

	return s.store.UpdateNote(ctx, noteID, update)
}

func (s *NoteService) DeleteNote(ctx context.Context, userID int64, noteID int64) error {
	// Get note and verify access
	n, err := s.store.GetNote(ctx, noteID)
	if err != nil {
		return err
	}

	if !s.isWorkspaceOwner(ctx, userID, n.WorkspaceID) {
		return errors.New("access denied")
	}

	return s.store.DeleteNote(ctx, noteID)
}

func (s *NoteService) isWorkspaceOwner(ctx context.Context, userID, workspaceID int64) bool {
	w, err := s.store.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return false
	}
	return w.UserID == userID
}

// sanitizeHTML removes dangerous HTML to prevent XSS.
// It strips script tags, event handlers, and javascript: URLs.
func sanitizeHTML(input string) string {
	// Remove script tags and their contents
	scriptRe := regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)
	output := scriptRe.ReplaceAllString(input, "")

	// Remove dangerous tags (iframe, object, embed, form, input)
	for _, tag := range []string{"iframe", "object", "embed", "form", "input"} {
		openRe := regexp.MustCompile(`(?i)<` + tag + `[^>]*>[\s\S]*?</` + tag + `>`)
		output = openRe.ReplaceAllString(output, "")
		selfCloseRe := regexp.MustCompile(`(?i)<` + tag + `[^>]*/?>`)
		output = selfCloseRe.ReplaceAllString(output, "")
	}

	// Remove on* event handlers from attributes (double quotes, single quotes, or unquoted)
	onEventRe := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*"[^"]*"`)
	output = onEventRe.ReplaceAllString(output, "")

	onEventSingleRe := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*'[^']*'`)
	output = onEventSingleRe.ReplaceAllString(output, "")

	onEventUnquotedRe := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*[^\s>]+`)
	output = onEventUnquotedRe.ReplaceAllString(output, "")

	// Remove javascript: URLs
	jsURLRe := regexp.MustCompile(`(?i)javascript:\s*[^"']*`)
	output = jsURLRe.ReplaceAllString(output, "")

	return output
}
