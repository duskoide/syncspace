package service

import (
	"context"
	"fmt"
	"strings"

	"syncspace/backend/internal/models"
)

func (s *Service) CreateBoard(ctx context.Context, moderatorID int64, req models.Board) (models.Board, error) {
	if strings.TrimSpace(req.Name) == "" {
		return models.Board{}, fmt.Errorf("board name is required")
	}
	req.ModeratorID = moderatorID
	return s.store.CreateBoard(ctx, req)
}

func (s *Service) GetBoard(ctx context.Context, id int64) (models.Board, error) {
	return s.store.GetBoard(ctx, id)
}

func (s *Service) ListBoards(ctx context.Context, userID int64, role string) ([]models.Board, error) {
	// Everyone can see all boards (discovery model)
	return s.store.ListBoards(ctx, 0)
}

func (s *Service) ListAllBoards(ctx context.Context) ([]models.Board, error) {
	return s.store.ListBoards(ctx, 0)
}

func (s *Service) UpdateBoard(ctx context.Context, userID, id int64, role string, req models.Board) (models.Board, error) {
	b, err := s.store.GetBoard(ctx, id)
	if err != nil {
		return models.Board{}, err
	}
	// Superadmin can update any board, moderator can only update their own
	if role != "superadmin" && b.ModeratorID != userID {
		return models.Board{}, fmt.Errorf("not authorized to update this board")
	}
	req.ModeratorID = b.ModeratorID
	return s.store.UpdateBoard(ctx, id, req)
}

func (s *Service) DeleteBoard(ctx context.Context, userID, id int64, role string) error {
	b, err := s.store.GetBoard(ctx, id)
	if err != nil {
		return err
	}
	// Superadmin can delete any board, moderator can only delete their own
	if role != "superadmin" && b.ModeratorID != userID {
		return fmt.Errorf("not authorized to delete this board")
	}
	return s.store.DeleteBoard(ctx, id)
}

// BoardMembership methods

func (s *Service) JoinBoard(ctx context.Context, userID, boardID int64, role string) (models.BoardMembership, error) {
	// Check if already a member
	existing, err := s.store.GetBoardMembershipByBoardAndUser(ctx, boardID, userID)
	if err == nil && existing.ID > 0 {
		return models.BoardMembership{}, fmt.Errorf("already a member of this board")
	}

	// Get board to check visibility
	board, err := s.store.GetBoard(ctx, boardID)
	if err != nil {
		return models.BoardMembership{}, fmt.Errorf("board not found")
	}

	if role == "" {
		role = "viewer"
	}

	bm := models.BoardMembership{
		BoardID: boardID,
		UserID:  userID,
		Role:    role,
	}

	// For private boards, you might want to set status to "pending" and require approval
	// For now, all joins are auto-approved regardless of visibility
	_ = board // Use board variable to avoid unused error

	return s.store.CreateBoardMembership(ctx, bm)
}

func (s *Service) UpdateMemberRole(ctx context.Context, moderatorID, membershipID int64, role string) error {
	bm, err := s.store.GetBoardMembership(ctx, membershipID)
	if err != nil {
		return err
	}
	b, err := s.store.GetBoard(ctx, bm.BoardID)
	if err != nil {
		return err
	}
	if b.ModeratorID != moderatorID {
		return fmt.Errorf("not authorized to update memberships for this board")
	}
	return s.store.UpdateBoardMembershipRole(ctx, membershipID, role)
}

func (s *Service) ListBoardMemberships(ctx context.Context, moderatorID, boardID int64) ([]models.BoardMembership, error) {
	b, err := s.store.GetBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if b.ModeratorID != moderatorID {
		return nil, fmt.Errorf("not authorized")
	}
	return s.store.ListBoardMembershipsByBoard(ctx, boardID)
}

func (s *Service) RemoveMember(ctx context.Context, moderatorID, boardID, userID int64) error {
	b, err := s.store.GetBoard(ctx, boardID)
	if err != nil {
		return err
	}
	if b.ModeratorID != moderatorID {
		return fmt.Errorf("not authorized")
	}
	return s.store.DeleteBoardMembership(ctx, boardID, userID)
}

func (s *Service) LeaveBoard(ctx context.Context, userID, boardID int64) error {
	return s.store.DeleteBoardMembership(ctx, boardID, userID)
}

// TextElement methods

func (s *Service) CreateTextElement(ctx context.Context, userID int64, req models.TextElement) (models.TextElement, error) {
	// Check if user has access to the board
	if !s.canAccessBoard(ctx, userID, req.BoardID) {
		return models.TextElement{}, fmt.Errorf("not authorized to add elements to this board")
	}
	
	// Check if user can edit (moderators, editors, or board creator)
	if !s.canEditBoard(ctx, userID, req.BoardID) {
		return models.TextElement{}, fmt.Errorf("not authorized to edit this board")
	}

	req.CreatedBy = userID
	
	// Set default dimensions if not provided
	if req.Width == 0 {
		req.Width = 200
	}
	if req.Height == 0 {
		req.Height = 150
	}
	if req.Color == "" {
		req.Color = "#FFFF88"
	}
	
	return s.store.CreateTextElement(ctx, req)
}

func (s *Service) GetTextElement(ctx context.Context, userID, id int64) (models.TextElement, error) {
	te, err := s.store.GetTextElement(ctx, id)
	if err != nil {
		return models.TextElement{}, err
	}
	
	if !s.canAccessBoard(ctx, userID, te.BoardID) {
		return models.TextElement{}, fmt.Errorf("not authorized")
	}
	
	return te, nil
}

func (s *Service) ListTextElementsByBoard(ctx context.Context, userID, boardID int64) ([]models.TextElement, error) {
	if !s.canAccessBoard(ctx, userID, boardID) {
		return nil, fmt.Errorf("not authorized to access this board")
	}
	return s.store.ListTextElementsByBoard(ctx, boardID)
}

func (s *Service) UpdateTextElement(ctx context.Context, userID, id int64, req models.TextElement) (models.TextElement, error) {
	te, err := s.store.GetTextElement(ctx, id)
	if err != nil {
		return models.TextElement{}, err
	}
	
	if !s.canEditBoard(ctx, userID, te.BoardID) {
		return models.TextElement{}, fmt.Errorf("not authorized to edit this board")
	}
	
	req.CreatedBy = te.CreatedBy
	return s.store.UpdateTextElement(ctx, id, req)
}

func (s *Service) DeleteTextElement(ctx context.Context, userID, id int64) error {
	te, err := s.store.GetTextElement(ctx, id)
	if err != nil {
		return err
	}
	
	// Only creator, moderator, or editors can delete
	if te.CreatedBy != userID && !s.canEditBoard(ctx, userID, te.BoardID) {
		return fmt.Errorf("not authorized to delete this element")
	}
	
	return s.store.DeleteTextElement(ctx, id)
}

func (s *Service) canAccessBoard(ctx context.Context, userID, boardID int64) bool {
	// Get board to check if user is moderator
	b, err := s.store.GetBoard(ctx, boardID)
	if err != nil {
		return false
	}
	if b.ModeratorID == userID {
		return true
	}
	
	// Check if user is a member
	_, err = s.store.GetBoardMembershipByBoardAndUser(ctx, boardID, userID)
	return err == nil
}

func (s *Service) canEditBoard(ctx context.Context, userID, boardID int64) bool {
	// Get board to check if user is moderator
	b, err := s.store.GetBoard(ctx, boardID)
	if err != nil {
		return false
	}
	if b.ModeratorID == userID {
		return true
	}
	
	// Check if user is an editor member
	bm, err := s.store.GetBoardMembershipByBoardAndUser(ctx, boardID, userID)
	if err != nil {
		return false
	}
	return bm.Role == "editor"
}

// Discussion methods

func (s *Service) CreateDiscussion(ctx context.Context, userID int64, req models.Discussion) (models.Discussion, error) {
	if strings.TrimSpace(req.Message) == "" {
		return models.Discussion{}, fmt.Errorf("message is required")
	}
	
	if !s.canAccessBoard(ctx, userID, req.BoardID) {
		return models.Discussion{}, fmt.Errorf("not authorized to post in this board")
	}
	
	req.UserID = userID
	return s.store.CreateDiscussion(ctx, req)
}

func (s *Service) GetDiscussion(ctx context.Context, userID, id int64) (models.Discussion, error) {
	d, err := s.store.GetDiscussion(ctx, id)
	if err != nil {
		return models.Discussion{}, err
	}
	
	if !s.canAccessBoard(ctx, userID, d.BoardID) {
		return models.Discussion{}, fmt.Errorf("not authorized")
	}
	
	return d, nil
}

func (s *Service) ListDiscussionsByBoard(ctx context.Context, userID, boardID int64, limit, offset int) ([]models.Discussion, error) {
	if !s.canAccessBoard(ctx, userID, boardID) {
		return nil, fmt.Errorf("not authorized to access this board")
	}
	
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.store.ListDiscussionsByBoard(ctx, boardID, limit, offset)
}

func (s *Service) ListDiscussionReplies(ctx context.Context, userID, parentID int64) ([]models.Discussion, error) {
	// First get the parent to check board access
	parent, err := s.store.GetDiscussion(ctx, parentID)
	if err != nil {
		return nil, err
	}
	
	if !s.canAccessBoard(ctx, userID, parent.BoardID) {
		return nil, fmt.Errorf("not authorized")
	}
	
	return s.store.ListDiscussionReplies(ctx, parentID)
}

func (s *Service) DeleteDiscussion(ctx context.Context, userID, id int64) error {
	d, err := s.store.GetDiscussion(ctx, id)
	if err != nil {
		return err
	}
	
	// Only creator or moderator can delete
	if d.UserID != userID {
		b, err := s.store.GetBoard(ctx, d.BoardID)
		if err != nil {
			return err
		}
		if b.ModeratorID != userID {
			return fmt.Errorf("not authorized")
		}
	}
	
	return s.store.DeleteDiscussion(ctx, id)
}
