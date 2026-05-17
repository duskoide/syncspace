package service

import (
	"context"
	"path/filepath"
	"testing"

	"syncspace/backend/internal/models"
	"syncspace/backend/internal/store"
)

func TestCreateTaskValidation(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	st, err := store.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	svc := New(st)
	_, err = svc.CreateTask(context.Background(), models.Task{Title: ""})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestNoteEnrichmentFlow(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	st, err := store.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	svc := New(st)
	n, err := svc.CreateNote(context.Background(), models.Note{Title: "n1", Content: "base"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.EnrichNote(context.Background(), n.ID, "")
	if err == nil {
		t.Fatal("expected error for empty topic")
	}
}
