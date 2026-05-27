package service

import (
	"context"
	"strings"
	"testing"

	"syncspace/backend/internal/models"
)

func createWorkspace(t *testing.T, svc *Service, userID int64, name string) models.Workspace {
	t.Helper()
	ws, err := svc.WorkspaceService.CreateWorkspace(context.Background(), userID, models.CreateWorkspaceRequest{
		Name: name,
	})
	if err != nil {
		t.Fatalf("createWorkspace(%s): %v", name, err)
	}
	return ws
}

// ==================== Create Note ====================

func TestCreateNote(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "note1@example.com", "password123", "Note1", "user")
	ws := createWorkspace(t, svc, user.ID, "Note Workspace")

	note, err := svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "Test Note",
	})
	if err != nil {
		t.Fatal(err)
	}
	if note.Title != "Test Note" {
		t.Fatalf("expected title 'Test Note', got %s", note.Title)
	}
	if note.Content != "" {
		t.Fatalf("expected empty content, got %s", note.Content)
	}
	if note.CreatedBy != user.ID {
		t.Fatalf("expected created_by %d, got %d", user.ID, note.CreatedBy)
	}
}

func TestCreateNoteEmptyTitle(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "note2@example.com", "password123", "Note2", "user")
	ws := createWorkspace(t, svc, user.ID, "Note WS2")

	_, err := svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestCreateNoteNoWorkspace(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "note3@example.com", "password123", "Note3", "user")

	_, err := svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{
		WorkspaceID: 0,
		Title:       "No WS",
	})
	if err == nil {
		t.Fatal("expected error for zero workspace_id")
	}
}

func TestCreateNoteAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "noteown@example.com", "password123", "NoteOwn", "user")
	user2 := registerUser(t, svc, "noteoth@example.com", "password123", "NoteOth", "user")
	ws := createWorkspace(t, svc, user1.ID, "Owner Only")

	_, err := svc.NoteService.CreateNote(context.Background(), user2.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "Injected",
	})
	if err == nil {
		t.Fatal("expected access denied for non-owner creating note")
	}
}

// ==================== Get Note ====================

func TestGetNote(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "getnote@example.com", "password123", "GetNote", "user")
	ws := createWorkspace(t, svc, user.ID, "Get Note WS")
	note, _ := svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "Get Note",
	})

	got, err := svc.NoteService.GetNote(context.Background(), user.ID, note.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Get Note" {
		t.Fatalf("expected title 'Get Note', got %s", got.Title)
	}
}

func TestGetNoteAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "gn1@example.com", "password123", "GN1", "user")
	user2 := registerUser(t, svc, "gn2@example.com", "password123", "GN2", "user")
	ws := createWorkspace(t, svc, user1.ID, "GN1 WS")
	note, _ := svc.NoteService.CreateNote(context.Background(), user1.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "Private Note",
	})

	_, err := svc.NoteService.GetNote(context.Background(), user2.ID, note.ID)
	if err == nil {
		t.Fatal("expected access denied for non-owner getting note")
	}
}

func TestGetNoteNotFound(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "gnnf@example.com", "password123", "GNNF", "user")

	_, err := svc.NoteService.GetNote(context.Background(), user.ID, 999)
	if err == nil {
		t.Fatal("expected error for non-existent note")
	}
}

// ==================== List Notes ====================

func TestListNotesByWorkspace(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "listnote@example.com", "password123", "ListNote", "user")
	ws := createWorkspace(t, svc, user.ID, "List WS")

	svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{WorkspaceID: ws.ID, Title: "Note1"})
	svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{WorkspaceID: ws.ID, Title: "Note2"})

	notes, err := svc.NoteService.ListNotesByWorkspace(context.Background(), user.ID, ws.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
}

func TestListNotesAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "ln1@example.com", "password123", "LN1", "user")
	user2 := registerUser(t, svc, "ln2@example.com", "password123", "LN2", "user")
	ws := createWorkspace(t, svc, user1.ID, "LN1 WS")
	svc.NoteService.CreateNote(context.Background(), user1.ID, models.CreateNoteRequest{WorkspaceID: ws.ID, Title: "Private"})

	_, err := svc.NoteService.ListNotesByWorkspace(context.Background(), user2.ID, ws.ID)
	if err == nil {
		t.Fatal("expected access denied for non-owner listing notes")
	}
}

// ==================== Update Note ====================

func TestUpdateNote(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "updnote@example.com", "password123", "UpdNote", "user")
	ws := createWorkspace(t, svc, user.ID, "Upd WS")
	note, _ := svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "Original",
	})

	updated, err := svc.NoteService.UpdateNote(context.Background(), user.ID, note.ID, models.UpdateNoteRequest{
		Title:   "Updated Title",
		Content: "<p>Updated content</p>",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Updated Title" {
		t.Fatalf("expected title 'Updated Title', got %s", updated.Title)
	}
	if updated.Content != "<p>Updated content</p>" {
		t.Fatalf("expected content '<p>Updated content</p>', got %s", updated.Content)
	}
}

func TestUpdateNoteAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "un1@example.com", "password123", "UN1", "user")
	user2 := registerUser(t, svc, "un2@example.com", "password123", "UN2", "user")
	ws := createWorkspace(t, svc, user1.ID, "UN1 WS")
	note, _ := svc.NoteService.CreateNote(context.Background(), user1.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "Protected",
	})

	_, err := svc.NoteService.UpdateNote(context.Background(), user2.ID, note.ID, models.UpdateNoteRequest{
		Title:   "Hacked",
		Content: "pwned",
	})
	if err == nil {
		t.Fatal("expected access denied for non-owner updating note")
	}
}

func TestUpdateNoteEmptyTitle(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "unempty@example.com", "password123", "UNEmpty", "user")
	ws := createWorkspace(t, svc, user.ID, "UNEmpty WS")
	note, _ := svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "Title",
	})

	_, err := svc.NoteService.UpdateNote(context.Background(), user.ID, note.ID, models.UpdateNoteRequest{
		Title: "",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

// ==================== Delete Note ====================

func TestDeleteNote(t *testing.T) {
	svc := setupTestService(t)
	user := registerUser(t, svc, "delnote@example.com", "password123", "DelNote", "user")
	ws := createWorkspace(t, svc, user.ID, "Del WS")
	note, _ := svc.NoteService.CreateNote(context.Background(), user.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "To Delete",
	})

	err := svc.NoteService.DeleteNote(context.Background(), user.ID, note.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.NoteService.GetNote(context.Background(), user.ID, note.ID)
	if err == nil {
		t.Fatal("expected error for deleted note")
	}
}

func TestDeleteNoteAccessDenied(t *testing.T) {
	svc := setupTestService(t)
	user1 := registerUser(t, svc, "dn1@example.com", "password123", "DN1", "user")
	user2 := registerUser(t, svc, "dn2@example.com", "password123", "DN2", "user")
	ws := createWorkspace(t, svc, user1.ID, "DN1 WS")
	note, _ := svc.NoteService.CreateNote(context.Background(), user1.ID, models.CreateNoteRequest{
		WorkspaceID: ws.ID,
		Title:       "Protected",
	})

	err := svc.NoteService.DeleteNote(context.Background(), user2.ID, note.ID)
	if err == nil {
		t.Fatal("expected access denied for non-owner deleting note")
	}
}

// ==================== HTML Sanitization ====================

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"safe html", "<p>Hello world</p>", "<p>Hello world</p>"},
		{"script tag", "<p>Hello</p><script>alert('xss')</script><p>World</p>", "<p>Hello</p><p>World</p>"},
		{"script tag uppercase", "<SCRIPT>alert('xss')</SCRIPT>", ""},
		{"iframe", "<iframe src='evil.com'></iframe>", ""},
		{"object", "<object data='evil.swf'></object>", ""},
		{"embed", "<embed src='evil.swf'>", ""},
		{"form", "<form action='evil.com'></form>", ""},
		{"input", "<input type='text' value='test'>", ""},
		{"onclick handler", `<div onclick="alert('xss')">`, `<div>`},
		{"onload handler", `<img src="x" onerror="alert('xss')">`, `<img src="x">`},
		{"onmouseover handler", `<div onmouseover="alert('xss')">`, `<div>`},
		{"javascript url", `<a href="javascript:alert('xss')">`, `<a href="">`},
		{"javascript url uppercase", `<A HREF="JAVASCRIPT:alert('xss')">`, `<A HREF="">`},
		{"nested script", `<div><script>document.cookie</script></div>`, `<div></div>`},
		{"script with content", `<script type="text/javascript">var x=1;</script>`, ""},
		{"event handler single quotes", `<div onfocus='alert(1)'>`, `<div>`},
		{"event handler no quotes", `<div onfocus=alert(1)>`, `<div>`},
		{"multiple attacks", `<img src="x" onerror="alert(1)"><script>alert(2)</script><a href="javascript:void(0)">`, `<img src="x"><a href="">`},
		{"valid attributes preserved", `<p class="text" id="main">Hello</p>`, `<p class="text" id="main">Hello</p>`},
		{"nested tags", `<div><p>Hello</p></div>`, `<div><p>Hello</p></div>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeHTML(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeHTML(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeHTMLPreservesContent(t *testing.T) {
	input := `<h1>Title</h1><p>This is <strong>bold</strong> and <em>italic</em> text.</p><ul><li>Item 1</li><li>Item 2</li></ul>`
	result := sanitizeHTML(input)

	if !strings.Contains(result, "<h1>Title</h1>") {
		t.Error("expected h1 to be preserved")
	}
	if !strings.Contains(result, "<strong>") {
		t.Error("expected strong tag to be preserved")
	}
	if !strings.Contains(result, "<em>") {
		t.Error("expected em tag to be preserved")
	}
	if !strings.Contains(result, "<ul>") {
		t.Error("expected ul tag to be preserved")
	}
}
