package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"syncspace/backend/internal/models"
	"syncspace/backend/internal/store"
)

type Service struct {
	store  *store.Store
	client *http.Client
}

func New(st *store.Store) *Service {
	return &Service{store: st, client: &http.Client{Timeout: 10 * time.Second}}
}

func (s *Service) ListTasks(ctx context.Context) ([]models.Task, error) {
	return s.store.ListTasks(ctx)
}
func (s *Service) GetTask(ctx context.Context, id int64) (models.Task, error) {
	return s.store.GetTask(ctx, id)
}
func (s *Service) DeleteTask(ctx context.Context, id int64) error { return s.store.DeleteTask(ctx, id) }

func (s *Service) CreateTask(ctx context.Context, t models.Task) (models.Task, error) {
	if strings.TrimSpace(t.Title) == "" {
		return models.Task{}, fmt.Errorf("title is required")
	}
	if t.Status == "" {
		t.Status = "todo"
	}
	return s.store.CreateTask(ctx, t)
}

func (s *Service) UpdateTask(ctx context.Context, id int64, t models.Task) (models.Task, error) {
	if strings.TrimSpace(t.Title) == "" {
		return models.Task{}, fmt.Errorf("title is required")
	}
	if t.Status == "" {
		t.Status = "todo"
	}
	return s.store.UpdateTask(ctx, id, t)
}

func (s *Service) ListNotes(ctx context.Context) ([]models.Note, error) {
	return s.store.ListNotes(ctx)
}
func (s *Service) GetNote(ctx context.Context, id int64) (models.Note, error) {
	return s.store.GetNote(ctx, id)
}
func (s *Service) DeleteNote(ctx context.Context, id int64) error { return s.store.DeleteNote(ctx, id) }

func (s *Service) CreateNote(ctx context.Context, n models.Note) (models.Note, error) {
	if strings.TrimSpace(n.Title) == "" {
		return models.Note{}, fmt.Errorf("title is required")
	}
	return s.store.CreateNote(ctx, n)
}

func (s *Service) UpdateNote(ctx context.Context, id int64, n models.Note) (models.Note, error) {
	if strings.TrimSpace(n.Title) == "" {
		return models.Note{}, fmt.Errorf("title is required")
	}
	return s.store.UpdateNote(ctx, id, n)
}

type wikiResponse struct {
	Extract string `json:"extract"`
}

func (s *Service) WikiSummary(ctx context.Context, topic string) (string, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return "", fmt.Errorf("topic is required")
	}
	u := "https://en.wikipedia.org/api/rest_v1/page/summary/" + url.PathEscape(topic)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("wikipedia error: %s", strings.TrimSpace(string(b)))
	}
	var out wikiResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if strings.TrimSpace(out.Extract) == "" {
		return "", fmt.Errorf("no summary found")
	}
	return out.Extract, nil
}

func (s *Service) EnrichNote(ctx context.Context, noteID int64, topic string) (models.Note, error) {
	n, err := s.store.GetNote(ctx, noteID)
	if err != nil {
		return models.Note{}, err
	}
	summary, err := s.WikiSummary(ctx, topic)
	if err != nil {
		return models.Note{}, err
	}
	if n.Content != "" {
		n.Content += "\n\n"
	}
	n.Content += "[Wikipedia: " + strings.TrimSpace(topic) + "]\n" + summary
	return s.store.UpdateNote(ctx, noteID, n)
}
