package requestid_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	requestid "github.com/docplanner/requestid"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestCreateConfig_Defaults(t *testing.T) {
	t.Parallel()

	cfg := requestid.CreateConfig()

	if cfg.HeaderName != "X-Request-ID" {
		t.Errorf("HeaderName = %q, want %q", cfg.HeaderName, "X-Request-ID")
	}

	if !cfg.Enabled {
		t.Error("Enabled = false, want true")
	}
}

func TestAddsHeader_WhenMissing(t *testing.T) {
	t.Parallel()

	cfg := requestid.CreateConfig()

	var capturedHeader string

	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Request-ID")
	})

	handler, err := requestid.New(context.Background(), next, cfg, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	handler.ServeHTTP(rec, req)

	if capturedHeader == "" {
		t.Fatal("expected X-Request-ID to be set on request, got empty")
	}

	if !uuidRegex.MatchString(capturedHeader) {
		t.Errorf("X-Request-ID = %q, does not match UUIDv4 format", capturedHeader)
	}
}

func TestPreservesExistingHeader(t *testing.T) {
	t.Parallel()

	cfg := requestid.CreateConfig()

	var capturedHeader string

	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Request-ID")
	})

	handler, err := requestid.New(context.Background(), next, cfg, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.Header.Set("X-Request-ID", "existing-id-123")
	handler.ServeHTTP(rec, req)

	if capturedHeader != "existing-id-123" {
		t.Errorf("X-Request-ID = %q, want %q", capturedHeader, "existing-id-123")
	}

	if rec.Header().Get("X-Request-ID") != "" {
		t.Errorf("response should not have X-Request-ID when header was already present, got %q", rec.Header().Get("X-Request-ID"))
	}
}

func TestDisabled_DoesNotAddHeader(t *testing.T) {
	t.Parallel()

	cfg := &requestid.Config{
		HeaderName: "X-Request-ID",
		Enabled:    false,
	}

	var capturedHeader string

	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Request-ID")
	})

	handler, err := requestid.New(context.Background(), next, cfg, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	handler.ServeHTTP(rec, req)

	if capturedHeader != "" {
		t.Errorf("X-Request-ID = %q, want empty when disabled", capturedHeader)
	}

	if rec.Header().Get("X-Request-ID") != "" {
		t.Errorf("response X-Request-ID = %q, want empty when disabled", rec.Header().Get("X-Request-ID"))
	}
}

func TestCustomHeaderName(t *testing.T) {
	t.Parallel()

	cfg := &requestid.Config{
		HeaderName: "X-Trace-ID",
		Enabled:    true,
	}

	var capturedHeader string

	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Trace-Id")
	})

	handler, err := requestid.New(context.Background(), next, cfg, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	handler.ServeHTTP(rec, req)

	if capturedHeader == "" {
		t.Fatal("expected X-Trace-ID to be set on request, got empty")
	}

	if !uuidRegex.MatchString(capturedHeader) {
		t.Errorf("X-Trace-ID = %q, does not match UUIDv4 format", capturedHeader)
	}

	if rec.Header().Get("X-Trace-Id") == "" {
		t.Fatal("expected X-Trace-ID to be set on response, got empty")
	}
}

func TestHeaderSetOnBothRequestAndResponse(t *testing.T) {
	t.Parallel()

	cfg := requestid.CreateConfig()

	var requestHeaderValue string

	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		requestHeaderValue = r.Header.Get("X-Request-ID")
	})

	handler, err := requestid.New(context.Background(), next, cfg, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	handler.ServeHTTP(rec, req)

	responseHeaderValue := rec.Header().Get("X-Request-ID")

	if requestHeaderValue == "" {
		t.Fatal("expected X-Request-ID on request, got empty")
	}

	if responseHeaderValue == "" {
		t.Fatal("expected X-Request-ID on response, got empty")
	}

	if requestHeaderValue != responseHeaderValue {
		t.Errorf("request header %q != response header %q", requestHeaderValue, responseHeaderValue)
	}
}

func TestUniqueIDsPerRequest(t *testing.T) {
	t.Parallel()

	cfg := requestid.CreateConfig()

	var ids []string

	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ids = append(ids, r.Header.Get("X-Request-ID"))
	})

	handler, err := requestid.New(context.Background(), next, cfg, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for range 10 {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		handler.ServeHTTP(rec, req)
	}

	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate request ID: %q", id)
		}

		seen[id] = true
	}
}

func TestNextHandlerCalled(t *testing.T) {
	t.Parallel()

	cfg := requestid.CreateConfig()

	called := false

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})

	handler, err := requestid.New(context.Background(), next, cfg, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("next handler was not called")
	}
}
