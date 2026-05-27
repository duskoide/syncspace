package service

import (
	"context"
	"testing"

	"syncspace/backend/internal/models"
)

// ==================== Create Template ====================

func TestCreateWorkspaceTemplate(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "tplws@example.com", "password123", "TplWS", "creator")
	ws := createWorkspace(t, svc, user.ID, "Template WS")
	svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "WS Note 1",
	})
	svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "WS Note 2",
	})

	tpl, err := svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:        "workspace",
		SourceID:    ws.ID,
		Name:        "My WS Template",
		Description: "A template workspace",
		Visibility:  "link",
	})
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Type != "workspace" {
		t.Fatalf("expected type 'workspace', got %s", tpl.Type)
	}
	if tpl.Visibility != "link" {
		t.Fatalf("expected visibility 'link', got %s", tpl.Visibility)
	}
}

func TestCreateTemplateInvalidType(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "tplinv@example.com", "password123", "TplInv", "creator")

	_, err := svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "note",
		SourceID: 1,
		Name:     "Note Type",
	})
	if err == nil {
		t.Fatal("expected error for note type")
	}
}

func TestCreateTemplateEmptyName(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "tplemp@example.com", "password123", "TplEmp", "creator")
	ws := createWorkspace(t, svc, user.ID, "Tpl Emp WS")

	_, err := svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "",
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCreateTemplateAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "tplden1@example.com", "password123", "TplDen1", "creator")
	user2 := registerUser(t, svc, "tplden2@example.com", "password123", "TplDen2", "creator")
	ws := createWorkspace(t, svc, user1.ID, "Tpl Den WS")

	_, err := svc.TemplateService.CreateTemplate(context.Background(), user2.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Stolen",
	})
	if err == nil {
		t.Fatal("expected access denied for non-owner creating template")
	}
}

func TestCreateTemplateInvalidVisibility(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "tplvis@example.com", "password123", "TplVis", "creator")
	ws := createWorkspace(t, svc, user.ID, "Tpl Vis WS")

	tpl, err := svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:       "workspace",
		SourceID:   ws.ID,
		Name:       "Vis Template",
		Visibility: "invalid",
	})
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Visibility != "public" {
		t.Fatalf("expected visibility to default to 'public', got %s", tpl.Visibility)
	}
}

// ==================== Get Template ====================

func TestGetTemplate(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "gtpl@example.com", "password123", "GTpl", "creator")
	ws := createWorkspace(t, svc, user.ID, "GTpl WS")

	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Get Template",
	})

	got, err := svc.TemplateService.GetTemplate(context.Background(), user.ID, tpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Get Template" {
		t.Fatalf("expected name 'Get Template', got %s", got.Name)
	}
}

func TestGetHiddenTemplateCreatorAccess(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "hiddpl@example.com", "password123", "HiddenTpl", "creator")
	admin, _ := svc.GetUserByID(context.Background(), 1)
	ws := createWorkspace(t, svc, user.ID, "Hidden WS")

	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Hidden Template",
	})
	svc.TemplateService.SetTemplateHidden(context.Background(), admin.ID, tpl.ID, true)

	got, err := svc.TemplateService.GetTemplate(context.Background(), user.ID, tpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsHidden {
		t.Fatal("expected template to be hidden")
	}
}

func TestGetHiddenTemplateNonCreatorDenied(t *testing.T) {
	svc := setupTestService(t)
	creator := registerUser(t, svc, "hiddcr@example.com", "password123", "HiddenCr", "creator")
	user := registerUser(t, svc, "hiddus@example.com", "password123", "HiddenUs", "user")
	admin, _ := svc.GetUserByID(context.Background(), 1)
	ws := createWorkspace(t, svc, creator.ID, "Hidden Cr WS")

	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), creator.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Hidden From Others",
	})
	svc.TemplateService.SetTemplateHidden(context.Background(), admin.ID, tpl.ID, true)

	_, err := svc.TemplateService.GetTemplate(context.Background(), user.ID, tpl.ID)
	if err == nil {
		t.Fatal("expected error for non-creator accessing hidden template")
	}
}

func TestGetHiddenTemplateSuperadminAccess(t *testing.T) {
	svc := setupTestService(t)
	creator := registerUser(t, svc, "hidsc@example.com", "password123", "HiddenSC", "creator")
	admin, _ := svc.GetUserByID(context.Background(), 1)
	ws := createWorkspace(t, svc, creator.ID, "Hidden SC WS")

	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), creator.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Hidden From Admin View",
	})
	svc.TemplateService.SetTemplateHidden(context.Background(), admin.ID, tpl.ID, true)

	got, err := svc.TemplateService.GetTemplate(context.Background(), admin.ID, tpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsHidden {
		t.Fatal("expected template to be hidden")
	}
}

// ==================== List Templates ====================

func TestListTemplates(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "ltpl@example.com", "password123", "Ltpl", "creator")
	ws := createWorkspace(t, svc, user.ID, "Ltpl WS")

	svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Public Template",
	})

	tpls, err := svc.TemplateService.ListTemplates(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tpl := range tpls {
		if tpl.Name == "Public Template" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected to find public template in list")
	}
}

func TestListTemplatesSearch(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "stpl@example.com", "password123", "Stpl", "creator")
	ws := createWorkspace(t, svc, user.ID, "Stpl WS")

	svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Unique Searchable Name",
	})

	tpls, err := svc.TemplateService.ListTemplates(context.Background(), "Unique")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tpl := range tpls {
		if tpl.Name == "Unique Searchable Name" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected to find template by search")
	}
}

func TestListMyTemplates(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "mtpl@example.com", "password123", "Mtpl", "creator")
	ws := createWorkspace(t, svc, user.ID, "Mtpl WS")

	svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "My Template",
	})

	tpls, err := svc.TemplateService.ListMyTemplates(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(tpls) != 1 {
		t.Fatalf("expected 1 template, got %d", len(tpls))
	}
}

// ==================== Update Template ====================

func TestUpdateTemplate(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "utpl@example.com", "password123", "UTpl", "creator")
	ws := createWorkspace(t, svc, user.ID, "UTpl WS")
	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Original Name",
	})

	updated, err := svc.TemplateService.UpdateTemplate(context.Background(), user.ID, tpl.ID, models.UpdateTemplateRequest{
		Name:        "Updated Name",
		Description: "Updated desc",
		Visibility:  "link",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Updated Name" {
		t.Fatalf("expected name 'Updated Name', got %s", updated.Name)
	}
	if updated.Visibility != "link" {
		t.Fatalf("expected visibility 'link', got %s", updated.Visibility)
	}
}

func TestUpdateTemplateAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "utp1@example.com", "password123", "UTp1", "creator")
	user2 := registerUser(t, svc, "utp2@example.com", "password123", "UTp2", "creator")
	ws := createWorkspace(t, svc, user1.ID, "UTp1 WS")
	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), user1.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Protected Template",
	})

	_, err := svc.TemplateService.UpdateTemplate(context.Background(), user2.ID, tpl.ID, models.UpdateTemplateRequest{
		Name: "Stolen",
	})
	if err == nil {
		t.Fatal("expected access denied for non-owner updating template")
	}
}

func TestUpdateTemplateEmptyName(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "utpemp@example.com", "password123", "UTpEmp", "creator")
	ws := createWorkspace(t, svc, user.ID, "UTpEmp WS")
	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Original",
	})

	_, err := svc.TemplateService.UpdateTemplate(context.Background(), user.ID, tpl.ID, models.UpdateTemplateRequest{
		Name: "",
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

// ==================== Delete Template ====================

func TestDeleteTemplate(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "dtpl@example.com", "password123", "DTpl", "creator")
	ws := createWorkspace(t, svc, user.ID, "DTpl WS")
	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "To Delete",
	})

	err := svc.TemplateService.DeleteTemplate(context.Background(), user.ID, tpl.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteTemplateAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "dtp1@example.com", "password123", "DTp1", "creator")
	user2 := registerUser(t, svc, "dtp2@example.com", "password123", "DTp2", "creator")
	ws := createWorkspace(t, svc, user1.ID, "DTp1 WS")
	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), user1.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Protected Delete",
	})

	err := svc.TemplateService.DeleteTemplate(context.Background(), user2.ID, tpl.ID)
	if err == nil {
		t.Fatal("expected access denied for non-owner deleting template")
	}
}

func TestDeleteTemplateSuperadminCanDelete(t *testing.T) {
	svc := setupTestService(t)
	creator := registerUser(t, svc, "dtpsc@example.com", "password123", "DTpSC", "creator")
	admin, _ := svc.GetUserByID(context.Background(), 1)
	ws := createWorkspace(t, svc, creator.ID, "DTpSC WS")
	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), creator.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Admin Delete",
	})

	err := svc.TemplateService.DeleteTemplate(context.Background(), admin.ID, tpl.ID)
	if err != nil {
		t.Fatal(err)
	}
}

// ==================== Set Template Hidden ====================

func TestSetTemplateHidden(t *testing.T) {
	svc := setupTestService(t)
	admin, _ := svc.GetUserByID(context.Background(), 1)
	creator := registerUser(t, svc, "seth@example.com", "password123", "SetH", "creator")
	ws := createWorkspace(t, svc, creator.ID, "SetH WS")
	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), creator.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "To Hide",
	})

	err := svc.TemplateService.SetTemplateHidden(context.Background(), admin.ID, tpl.ID, true)
	if err != nil {
		t.Fatal(err)
	}

	got, _ := svc.TemplateService.GetTemplate(context.Background(), admin.ID, tpl.ID)
	if !got.IsHidden {
		t.Fatal("expected template to be hidden")
	}
}

func TestSetTemplateHiddenNonAdminDenied(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "sethna@example.com", "password123", "SetHNA", "creator")
	ws := createWorkspace(t, svc, user.ID, "SetHNA WS")
	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Non Admin Hide",
	})

	err := svc.TemplateService.SetTemplateHidden(context.Background(), user.ID, tpl.ID, true)
	if err == nil {
		t.Fatal("expected access denied for non-admin hiding template")
	}
}

// ==================== Clone Template ====================

func TestCloneWorkspaceTemplate(t *testing.T) {
	svc := setupTestService(t)
	creator := registerUser(t, svc, "cws1@example.com", "password123", "Cws1", "creator")
	user := registerUser(t, svc, "cws2@example.com", "password123", "Cws2", "user")
	ws := createWorkspace(t, svc, creator.ID, "Clonable Workspace")
	svc.NoteService.CreateNote(context.Background(), creator.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "Original Note",
	})

	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), creator.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Clonable Workspace Template",
	})

	cloned, _, err := svc.TemplateService.CloneTemplate(context.Background(), user.ID, tpl.ID, models.CloneTemplateRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if cloned == nil {
		t.Fatal("expected cloned workspace")
	}
	if cloned.Name != "Clonable Workspace (Copy)" {
		t.Fatalf("expected name 'Clonable Workspace (Copy)', got %s", cloned.Name)
	}
	if cloned.UserID != user.ID {
		t.Fatalf("expected user_id %d, got %d", user.ID, cloned.UserID)
	}

	notes, _ := svc.NoteService.ListNotesByWorkspace(context.Background(), user.ID, cloned.ID)
	if len(notes) != 1 {
		t.Fatalf("expected 1 cloned note, got %d", len(notes))
	}
	if notes[0].Title != "Original Note" {
		t.Fatalf("expected cloned note title 'Original Note', got %s", notes[0].Title)
	}
}

func TestCloneHiddenTemplateDenied(t *testing.T) {
	svc := setupTestService(t)
	creator := registerUser(t, svc, "chd1@example.com", "password123", "CHD1", "creator")
	user := registerUser(t, svc, "chd2@example.com", "password123", "CHD2", "user")
	admin, _ := svc.GetUserByID(context.Background(), 1)
	ws := createWorkspace(t, svc, creator.ID, "CHD1 WS")
	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), creator.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Hidden Clone",
	})
	svc.TemplateService.SetTemplateHidden(context.Background(), admin.ID, tpl.ID, true)

	_, _, err := svc.TemplateService.CloneTemplate(context.Background(), user.ID, tpl.ID, models.CloneTemplateRequest{})
	if err == nil {
		t.Fatal("expected error for cloning hidden template")
	}
}

// ==================== Update Template Content ====================

func TestUpdateTemplateContent(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "utc1@example.com", "password123", "UTC1", "creator")
	ws := createWorkspace(t, svc, user.ID, "UTC1 WS")
	svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "Evolving Note",
	})

	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), user.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Evolving Template",
	})

	svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "New Note",
	})

	updated, err := svc.TemplateService.UpdateTemplateContent(context.Background(), user.ID, tpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ContentSnapshot == "" {
		t.Fatal("expected non-empty content snapshot after update")
	}
}

func TestUpdateTemplateContentAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "utc2@example.com", "password123", "UTC2", "creator")
	user2 := registerUser(t, svc, "utc3@example.com", "password123", "UTC3", "creator")
	ws := createWorkspace(t, svc, user1.ID, "UTC2 WS")

	tpl, _ := svc.TemplateService.CreateTemplate(context.Background(), user1.ID, models.CreateTemplateRequest{
		Type:     "workspace",
		SourceID: ws.ID,
		Name:     "Protected Template",
	})

	_, err := svc.TemplateService.UpdateTemplateContent(context.Background(), user2.ID, tpl.ID)
	if err == nil {
		t.Fatal("expected access denied for non-owner updating template content")
	}
}
