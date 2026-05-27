package service

import (
	"context"
	"testing"

	"syncspace/backend/internal/models"
)

// ==================== Create Workspace ====================

func TestCreateWorkspace(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "ws1@example.com", "password123", "WS1", "user")

	ws, err := svc.WorkspaceService.CreateWorkspace(context.Background(), user.ID, models.CreateWorkspaceRequest{
		Name:        "My Workspace",
		Description: "A workspace",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ws.Name != "My Workspace" {
		t.Fatalf("expected name 'My Workspace', got %s", ws.Name)
	}
	if ws.UserID != user.ID {
		t.Fatalf("expected user_id %d, got %d", user.ID, ws.UserID)
	}
}

func TestCreateWorkspaceEmptyName(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "ws2@example.com", "password123", "WS2", "user")

	_, err := svc.WorkspaceService.CreateWorkspace(context.Background(), user.ID, models.CreateWorkspaceRequest{
		Name: "",
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCreateWorkspaceWhitespaceName(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "ws3@example.com", "password123", "WS3", "user")

	_, err := svc.WorkspaceService.CreateWorkspace(context.Background(), user.ID, models.CreateWorkspaceRequest{
		Name: "   ",
	})
	if err == nil {
		t.Fatal("expected error for whitespace-only name")
	}
}

// ==================== Get Workspace ====================

func TestGetWorkspace(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "getws@example.com", "password123", "GetWS", "user")

	ws, _ := svc.WorkspaceService.CreateWorkspace(context.Background(), user.ID, models.CreateWorkspaceRequest{
		Name: "Get WS",
	})

	got, err := svc.WorkspaceService.GetWorkspace(context.Background(), user.ID, ws.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != ws.ID {
		t.Fatalf("expected ID %d, got %d", ws.ID, got.ID)
	}
}

func TestGetWorkspaceAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "owner@example.com", "password123", "Owner", "user")
	user2 := registerUser(t, svc, "other@example.com", "password123", "Other", "user")

	ws, _ := svc.WorkspaceService.CreateWorkspace(context.Background(), user1.ID, models.CreateWorkspaceRequest{
		Name: "Owner WS",
	})

	_, err := svc.WorkspaceService.GetWorkspace(context.Background(), user2.ID, ws.ID)
	if err == nil {
		t.Fatal("expected access denied for non-owner")
	}
}

func TestGetWorkspaceNotFound(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "notfound@example.com", "password123", "NotFound", "user")

	_, err := svc.WorkspaceService.GetWorkspace(context.Background(), user.ID, 999)
	if err == nil {
		t.Fatal("expected error for non-existent workspace")
	}
}

// ==================== List Workspaces ====================

func TestListWorkspaces(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "listws@example.com", "password123", "ListWS", "user")

	svc.WorkspaceService.CreateWorkspace(context.Background(), user.ID, models.CreateWorkspaceRequest{Name: "WS1"})
	svc.WorkspaceService.CreateWorkspace(context.Background(), user.ID, models.CreateWorkspaceRequest{Name: "WS2"})

	wsList, err := svc.WorkspaceService.ListWorkspaces(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(wsList) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(wsList))
	}
}

func TestListWorkspacesIsolation(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "iso1@example.com", "password123", "Iso1", "user")
	user2 := registerUser(t, svc, "iso2@example.com", "password123", "Iso2", "user")

	svc.WorkspaceService.CreateWorkspace(context.Background(), user1.ID, models.CreateWorkspaceRequest{Name: "User1 WS"})
	svc.WorkspaceService.CreateWorkspace(context.Background(), user2.ID, models.CreateWorkspaceRequest{Name: "User2 WS"})

	ws1, _ := svc.WorkspaceService.ListWorkspaces(context.Background(), user1.ID)
	ws2, _ := svc.WorkspaceService.ListWorkspaces(context.Background(), user2.ID)

	if len(ws1) != 1 || len(ws2) != 1 {
		t.Fatalf("expected 1 workspace each, got %d and %d", len(ws1), len(ws2))
	}
	if ws1[0].Name != "User1 WS" {
		t.Fatalf("expected 'User1 WS', got %s", ws1[0].Name)
	}
}

// ==================== Update Workspace ====================

func TestUpdateWorkspace(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "updatews@example.com", "password123", "UpdateWS", "user")

	ws, _ := svc.WorkspaceService.CreateWorkspace(context.Background(), user.ID, models.CreateWorkspaceRequest{
		Name: "Original",
	})

	updated, err := svc.WorkspaceService.UpdateWorkspace(context.Background(), user.ID, ws.ID, models.UpdateWorkspaceRequest{
		Name:        "Updated",
		Description: "New description",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Updated" {
		t.Fatalf("expected name 'Updated', got %s", updated.Name)
	}
	if updated.Description != "New description" {
		t.Fatalf("expected description 'New description', got %s", updated.Description)
	}
}

func TestUpdateWorkspaceAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "upd1@example.com", "password123", "Upd1", "user")
	user2 := registerUser(t, svc, "upd2@example.com", "password123", "Upd2", "user")

	ws, _ := svc.WorkspaceService.CreateWorkspace(context.Background(), user1.ID, models.CreateWorkspaceRequest{Name: "Owner"})

	_, err := svc.WorkspaceService.UpdateWorkspace(context.Background(), user2.ID, ws.ID, models.UpdateWorkspaceRequest{
		Name: "Hacked",
	})
	if err == nil {
		t.Fatal("expected access denied for non-owner update")
	}
}

func TestUpdateWorkspaceEmptyName(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "updempty@example.com", "password123", "UpdEmpty", "user")

	ws, _ := svc.WorkspaceService.CreateWorkspace(context.Background(), user.ID, models.CreateWorkspaceRequest{Name: "Original"})

	_, err := svc.WorkspaceService.UpdateWorkspace(context.Background(), user.ID, ws.ID, models.UpdateWorkspaceRequest{
		Name: "",
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

// ==================== Delete Workspace ====================

func TestDeleteWorkspace(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "delws@example.com", "password123", "DelWS", "user")

	ws, _ := svc.WorkspaceService.CreateWorkspace(context.Background(), user.ID, models.CreateWorkspaceRequest{
		Name: "To Delete",
	})

	err := svc.WorkspaceService.DeleteWorkspace(context.Background(), user.ID, ws.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.WorkspaceService.GetWorkspace(context.Background(), user.ID, ws.ID)
	if err == nil {
		t.Fatal("expected error for deleted workspace")
	}
}

func TestDeleteWorkspaceAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "del1@example.com", "password123", "Del1", "user")
	user2 := registerUser(t, svc, "del2@example.com", "password123", "Del2", "user")

	ws, _ := svc.WorkspaceService.CreateWorkspace(context.Background(), user1.ID, models.CreateWorkspaceRequest{Name: "Protected"})

	err := svc.WorkspaceService.DeleteWorkspace(context.Background(), user2.ID, ws.ID)
	if err == nil {
		t.Fatal("expected access denied for non-owner delete")
	}
}

// ==================== IsWorkspaceOwner ====================

func TestIsWorkspaceOwner(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "isowner@example.com", "password123", "IsOwner", "user")

	ws, _ := svc.WorkspaceService.CreateWorkspace(context.Background(), user.ID, models.CreateWorkspaceRequest{Name: "Owned"})

	if !svc.WorkspaceService.IsWorkspaceOwner(context.Background(), user.ID, ws.ID) {
		t.Fatal("expected user to be owner")
	}
}

func TestIsWorkspaceOwnerFalse(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "notown1@example.com", "password123", "NotOwn1", "user")
	user2 := registerUser(t, svc, "notown2@example.com", "password123", "NotOwn2", "user")

	ws, _ := svc.WorkspaceService.CreateWorkspace(context.Background(), user1.ID, models.CreateWorkspaceRequest{Name: "Not Owned"})

	if svc.WorkspaceService.IsWorkspaceOwner(context.Background(), user2.ID, ws.ID) {
		t.Fatal("expected user not to be owner")
	}
}
