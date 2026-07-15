package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouterLoadsTemplatesAndHealthRoute(t *testing.T) {
	router, err := NewRouter()
	if err != nil {
		t.Fatalf("create router: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"up"`) {
		t.Fatalf("unexpected health response: %s", recorder.Body.String())
	}
}
