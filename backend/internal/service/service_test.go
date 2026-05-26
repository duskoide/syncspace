package service

import (
	"context"
	"path/filepath"
	"testing"

	"syncspace/backend/internal/models"
	"syncspace/backend/internal/store"
)

func TestRegisterValidation(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	st, err := store.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	svc := New(st, t.TempDir())

	// Empty email should fail
	_, err = svc.Register(context.Background(), models.RegisterRequest{
		Email:    "",
		Password: "password123",
		Name:     "Test",
		Role:     "user",
	})
	if err == nil {
		t.Fatal("expected error for empty email")
	}

	// Short password should fail
	_, err = svc.Register(context.Background(), models.RegisterRequest{
		Email:    "test@example.com",
		Password: "short",
		Name:     "Test",
		Role:     "user",
	})
	if err == nil {
		t.Fatal("expected error for short password")
	}

	// Invalid role should fail
	_, err = svc.Register(context.Background(), models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test",
		Role:     "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
}

func TestWorkspaceCRUD(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	st, err := store.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	svc := New(st, t.TempDir())
	ctx := context.Background()

	// Create user first
	user, err := svc.Register(ctx, models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     "user",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create workspace
	ws, err := svc.WorkspaceService.CreateWorkspace(ctx, user.ID, models.CreateWorkspaceRequest{
		Name:        "Test Workspace",
		Description: "A test workspace",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ws.Name != "Test Workspace" {
		t.Fatalf("expected workspace name 'Test Workspace', got %s", ws.Name)
	}

	// List workspaces
	workspaces, err := svc.WorkspaceService.ListWorkspaces(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(workspaces))
	}

	// Update workspace
	updated, err := svc.WorkspaceService.UpdateWorkspace(ctx, user.ID, ws.ID, models.UpdateWorkspaceRequest{
		Name:        "Updated Workspace",
		Description: "Updated description",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Updated Workspace" {
		t.Fatalf("expected updated name 'Updated Workspace', got %s", updated.Name)
	}
}

func TestWikipediaSummary(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	st, err := store.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	svc := New(st, t.TempDir())
	ctx := context.Background()

	// Empty topic should fail
	_, err = svc.WikiSummary(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty topic")
	}

	// Valid topic should return summary
	summary, err := svc.WikiSummary(ctx, "Artificial_intelligence")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
}
