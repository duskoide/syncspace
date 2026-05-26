package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"syncspace/backend/internal/auth"
	"syncspace/backend/internal/models"
)

func (s *Service) Register(ctx context.Context, req models.RegisterRequest) (models.User, error) {
	if strings.TrimSpace(req.Email) == "" {
		return models.User{}, fmt.Errorf("email is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return models.User{}, fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return models.User{}, fmt.Errorf("password must be at least 8 characters")
	}
	if strings.TrimSpace(req.Name) == "" {
		return models.User{}, fmt.Errorf("name is required")
	}
	if req.Role != "creator" && req.Role != "user" {
		return models.User{}, fmt.Errorf("role must be creator or user")
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	u := models.User{
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: hash,
		Name:         strings.TrimSpace(req.Name),
		Role:         req.Role,
		Status:       "active", // Auto-approved on registration
	}

	return s.store.CreateUser(ctx, u)
}

func (s *Service) Login(ctx context.Context, req models.LoginRequest) (models.TokenPair, models.User, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return models.TokenPair{}, models.User{}, fmt.Errorf("email and password are required")
	}

	u, err := s.store.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if err == sql.ErrNoRows {
			return models.TokenPair{}, models.User{}, fmt.Errorf("invalid email or password")
		}
		return models.TokenPair{}, models.User{}, fmt.Errorf("database error: %w", err)
	}

	if u.Status == "suspended" {
		return models.TokenPair{}, models.User{}, fmt.Errorf("account suspended")
	}

	if !auth.CheckPassword(req.Password, u.PasswordHash) {
		return models.TokenPair{}, models.User{}, fmt.Errorf("invalid email or password")
	}

	token, err := auth.GenerateToken(u.ID, u.Email, u.Name, u.Role)
	if err != nil {
		return models.TokenPair{}, models.User{}, fmt.Errorf("failed to generate token: %w", err)
	}

	return models.TokenPair{
		AccessToken: token,
		ExpiresIn:   86400, // 24 hours
	}, u, nil
}

func (s *Service) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	return s.store.GetUserByID(ctx, id)
}

func (s *Service) ListUsers(ctx context.Context, role, status string) ([]models.User, error) {
	return s.store.ListUsers(ctx, role, status)
}

func (s *Service) ActivateUser(ctx context.Context, adminID, userID int64) error {
	admin, err := s.store.GetUserByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("admin not found")
	}
	if admin.Role != "superadmin" {
		return fmt.Errorf("insufficient permissions")
	}

	_, err = s.store.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	return s.store.UpdateUserStatus(ctx, userID, "active")
}

func (s *Service) SuspendUser(ctx context.Context, adminID, userID int64) error {
	admin, err := s.store.GetUserByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("admin not found")
	}
	if admin.Role != "superadmin" {
		return fmt.Errorf("insufficient permissions")
	}

	_, err = s.store.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	return s.store.UpdateUserStatus(ctx, userID, "suspended")
}

func (s *Service) DeleteUser(ctx context.Context, adminID, userID int64) error {
	admin, err := s.store.GetUserByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("admin not found")
	}
	if admin.Role != "superadmin" {
		return fmt.Errorf("insufficient permissions")
	}

	return s.store.DeleteUser(ctx, userID)
}
