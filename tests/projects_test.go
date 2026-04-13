package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestProjectsStatsRouteWorks(t *testing.T) {
	InitTestMode(t)
	router := NewRouter()

	projectID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/stats", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	rec := httptest.NewRecorder()

	router.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
