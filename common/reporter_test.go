package common

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestSetReporter_NilRemovesHook(t *testing.T) {
	defer SetReporter(nil)

	called := false
	SetReporter(func(context.Context, *AppError) { called = true })
	if Reporter() == nil {
		t.Fatal("Reporter should be installed")
	}

	SetReporter(nil)
	if Reporter() != nil {
		t.Fatal("Reporter should be removed")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	WriteJSON(rec, req, ErrInternal(errors.New("x")))
	if called {
		t.Fatal("removed reporter must not be called")
	}
}

func TestWriteJSON_InvokesReporter(t *testing.T) {
	defer SetReporter(nil)

	var (
		mu   sync.Mutex
		got  *AppError
		gctx context.Context
	)
	SetReporter(func(ctx context.Context, err *AppError) {
		mu.Lock()
		defer mu.Unlock()
		got = err
		gctx = ctx
	})

	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, r, ErrUnauthorized(errors.New("missing token")))
	}), RequestIDMiddleware)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secret", nil)
	req.Header.Set(RequestIDHeader, "rid-xyz")
	h.ServeHTTP(rec, req)

	mu.Lock()
	defer mu.Unlock()
	if got == nil {
		t.Fatal("reporter was not invoked")
	}
	if got.Code != CodeUnauthorized {
		t.Errorf("code: %q", got.Code)
	}
	if got.RequestID != "rid-xyz" {
		t.Errorf("request_id stamped before reporter: %q", got.RequestID)
	}
	if RequestIDFromContext(gctx) != "rid-xyz" {
		t.Errorf("context request id: %q", RequestIDFromContext(gctx))
	}
}

func TestReporter_PanicDoesNotBreakResponse(t *testing.T) {
	defer SetReporter(nil)
	SetReporter(func(context.Context, *AppError) { panic("reporter blew up") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	WriteJSON(rec, req, ErrInternal(errors.New("x")))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: %d", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("body should still be written even if reporter panics")
	}
}

func TestWriteJSON_RetryAfterHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	WriteJSON(rec, req, ErrRateLimited(42, nil))

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status: %d", rec.Code)
	}
	if got := rec.Header().Get("Retry-After"); got != "42" {
		t.Errorf("Retry-After: %q", got)
	}
}

func TestWriteJSON_NoRetryAfterForOtherErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	WriteJSON(rec, req, ErrEntityNotFound("user", nil))

	if got := rec.Header().Get("Retry-After"); got != "" {
		t.Errorf("Retry-After should be empty, got %q", got)
	}
}

func TestRetryAfterSeconds(t *testing.T) {
	cases := []struct {
		name string
		err  *AppError
		want int
	}{
		{"rate-limit constructor", ErrRateLimited(7, nil), 7},
		{"zero hint", ErrRateLimited(0, nil), 0},
		{"negative hint", ErrRateLimited(-5, nil), 0},
		{"no details", ErrInternal(errors.New("x")), 0},
		{"nil receiver", nil, 0},
		{"float in details", New(429, "X", "m", nil).WithDetails(map[string]any{"retry_after_seconds": float64(12)}), 12},
		{"int64 in details", New(429, "X", "m", nil).WithDetails(map[string]any{"retry_after_seconds": int64(33)}), 33},
		{"non-numeric details", New(429, "X", "m", nil).WithDetails(map[string]any{"retry_after_seconds": "soon"}), 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.err.RetryAfterSeconds(); got != c.want {
				t.Errorf("got %d want %d", got, c.want)
			}
		})
	}
}
