package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestTasksListUsesLimitPagination(t *testing.T) {
	InitTestMode(t)
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+uuid.New().String()+"/tasks?page=2&limit=7", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	rec := httptest.NewRecorder()

	router.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if router.Tasks.LastPagination.Page != 2 || router.Tasks.LastPagination.PageSize != 7 {
		t.Fatalf("expected pagination page=2 limit=7, got page=%d limit=%d", router.Tasks.LastPagination.Page, router.Tasks.LastPagination.PageSize)
	}
}
