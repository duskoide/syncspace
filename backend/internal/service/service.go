package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"syncspace/backend/internal/store"
)

type Service struct {
	store            *store.Store
	client           *http.Client
	WorkspaceService *WorkspaceService
	NoteService      *NoteService
	TemplateService  *TemplateService
	FileService      *FileService
}

func New(st *store.Store, uploadDir string) *Service {
	s := &Service{
		store:  st,
		client: &http.Client{Timeout: 10 * time.Second},
	}
	// Initialize sub-services
	s.WorkspaceService = &WorkspaceService{store: st}
	s.NoteService = &NoteService{store: st}
	s.TemplateService = &TemplateService{store: st}
	s.FileService = NewFileService(st, uploadDir)
	return s
}

// ==================== Wikipedia Integration ====================

type wikiResponse struct {
	Extract string `json:"extract"`
}

type UpstreamError struct {
	Message string
}

func (e *UpstreamError) Error() string {
	return e.Message
}

func IsUpstreamError(err error) bool {
	var upstreamErr *UpstreamError
	return errors.As(err, &upstreamErr)
}

func (s *Service) WikiSummary(ctx context.Context, topic string) (string, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return "", fmt.Errorf("topic is required")
	}
	u := "https://en.wikipedia.org/api/rest_v1/page/summary/" + url.PathEscape(topic)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "SyncSpace/1.0 (Note-taking app)")
	req.Header.Set("Api-User-Agent", "SyncSpace/1.0 (Note-taking app)")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", &UpstreamError{Message: fmt.Sprintf("wikipedia request failed: %v", err)}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", &UpstreamError{Message: fmt.Sprintf("wikipedia error: %s", strings.TrimSpace(string(b)))}
	}
	var out wikiResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", &UpstreamError{Message: fmt.Sprintf("wikipedia decode failed: %v", err)}
	}
	if strings.TrimSpace(out.Extract) == "" {
		return "", &UpstreamError{Message: "no summary found"}
	}
	return out.Extract, nil
}
