package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	t.Run("returns root cause when present", func(t *testing.T) {
		root := errors.New("db is down")
		appErr := New(500, "X", "msg", root)
		if appErr.Error() != "db is down" {
			t.Fatalf("got %q", appErr.Error())
		}
	})
	t.Run("falls back to message when no root", func(t *testing.T) {
		appErr := New(400, "X", "bad", nil)
		if appErr.Error() != "bad" {
			t.Fatalf("got %q", appErr.Error())
		}
	})
	t.Run("nil receiver is safe", func(t *testing.T) {
		var e *AppError
		if got := e.Error(); got != "" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestAppError_UnwrapAndIs(t *testing.T) {
	root := ErrRecordNotFound
	wrapped := ErrEntityNotFound("user", fmt.Errorf("repo: %w", root))
	if !errors.Is(wrapped, ErrRecordNotFound) {
		t.Fatal("errors.Is should match through the wrap chain")
	}
}

func TestNew_LogLineDefaults(t *testing.T) {
	t.Run("log defaults to root.Error", func(t *testing.T) {
		root := errors.New("syntax error at column 12")
		e := New(400, "X", "Invalid request", root)
		if e.LogLine() != "syntax error at column 12" {
			t.Fatalf("got %q", e.LogLine())
		}
	})
	t.Run("log defaults to message when no root", func(t *testing.T) {
		e := New(400, "X", "bad", nil)
		if e.LogLine() != "bad" {
			t.Fatalf("got %q", e.LogLine())
		}
	})
	t.Run("WithLog overrides", func(t *testing.T) {
		e := New(400, "X", "bad", nil).WithLog("rich log detail")
		if e.LogLine() != "rich log detail" {
			t.Fatalf("got %q", e.LogLine())
		}
	})
}

func TestConstructors_StatusAndCode(t *testing.T) {
	cases := []struct {
		name    string
		err     *AppError
		wantSt  int
		wantCd  string
	}{
		{"NotFound", ErrEntityNotFound("user", nil), http.StatusNotFound, CodeEntityNotFound},
		{"AlreadyExists", ErrEntityAlreadyExists("user", nil), http.StatusConflict, CodeEntityAlreadyExists},
		{"Create", ErrCannotCreateEntity("user", nil), http.StatusUnprocessableEntity, CodeEntityCreateFailed},
		{"Update", ErrCannotUpdateEntity("user", nil), http.StatusUnprocessableEntity, CodeEntityUpdateFailed},
		{"Delete", ErrCannotDeleteEntity("user", nil), http.StatusUnprocessableEntity, CodeEntityDeleteFailed},
		{"Get", ErrCannotGetEntity("user", nil), http.StatusUnprocessableEntity, CodeEntityGetFailed},
		{"List", ErrCannotListEntity("user", nil), http.StatusUnprocessableEntity, CodeEntityListFailed},
		{"Unauthorized", ErrUnauthorized(nil), http.StatusUnauthorized, CodeUnauthorized},
		{"InvalidCreds", ErrInvalidCredentials(nil), http.StatusUnauthorized, CodeInvalidCredentials},
		{"TokenExpired", ErrTokenExpired(nil), http.StatusUnauthorized, CodeTokenExpired},
		{"Forbidden", ErrForbidden(nil), http.StatusForbidden, CodeForbidden},
		{"InvalidReq", ErrInvalidRequest(nil), http.StatusBadRequest, CodeInvalidRequest},
		{"Validation", ErrValidation(nil, nil), http.StatusUnprocessableEntity, CodeValidationFailed},
		{"WeakPwd", ErrWeakPassword(nil), http.StatusUnprocessableEntity, CodeWeakPassword},
		{"RateLimited", ErrRateLimited(30, nil), http.StatusTooManyRequests, CodeRateLimited},
		{"Conflict", ErrConflict("", nil), http.StatusConflict, CodeConflict},
		{"UnsupportedMedia", ErrUnsupportedMediaType(nil), http.StatusUnsupportedMediaType, CodeUnsupportedMedia},
		{"PayloadTooLarge", ErrPayloadTooLarge(nil), http.StatusRequestEntityTooLarge, CodePayloadTooLarge},
		{"DB", ErrDB(errors.New("x")), http.StatusInternalServerError, CodeDatabaseError},
		{"Internal", ErrInternal(errors.New("x")), http.StatusInternalServerError, CodeInternal},
		{"Upstream", ErrUpstream("billing", nil), http.StatusBadGateway, CodeUpstreamError},
		{"Timeout", ErrTimeout("db.query", nil), http.StatusGatewayTimeout, CodeTimeout},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.err.StatusCode != c.wantSt {
				t.Errorf("status: got %d want %d", c.err.StatusCode, c.wantSt)
			}
			if c.err.Code != c.wantCd {
				t.Errorf("code: got %q want %q", c.err.Code, c.wantCd)
			}
		})
	}
}

func TestEntityNormalisation(t *testing.T) {
	e := ErrEntityNotFound("  User  ", nil)
	if !strings.Contains(e.Message, "user not found") {
		t.Fatalf("message %q should be lowercased and trimmed", e.Message)
	}
	d, ok := e.Details.(EntityDetails)
	if !ok {
		t.Fatalf("details type: %T", e.Details)
	}
	if d.Entity != "user" {
		t.Fatalf("entity: %q", d.Entity)
	}
}

func TestValidation_DetailsShape(t *testing.T) {
	fields := []FieldError{
		{Field: "email", Message: "must be valid email", Code: "INVALID_EMAIL"},
		{Field: "password", Message: "too short"},
	}
	e := ErrValidation(fields, nil)
	m, ok := e.Details.(map[string]any)
	if !ok {
		t.Fatalf("details type: %T", e.Details)
	}
	if _, ok := m["fields"]; !ok {
		t.Fatal("details should carry fields key")
	}
	body, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"field":"email"`) {
		t.Fatalf("serialized body missing field: %s", body)
	}
}

func TestJSON_DoesNotLeakInternalFields(t *testing.T) {
	root := errors.New("postgres: relation users does not exist")
	e := ErrDB(root)
	body, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if strings.Contains(s, "postgres") {
		t.Fatalf("raw root error leaked into JSON: %s", s)
	}
	if strings.Contains(s, "log") {
		t.Fatalf("log field leaked into JSON: %s", s)
	}
}

func TestWriteJSON_AndRequestID(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, r, ErrEntityNotFound("user", ErrRecordNotFound))
	}), RequestIDMiddleware)

	t.Run("generates request id when client omits it", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("status: %d", w.Code)
		}
		var body AppError
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Code != CodeEntityNotFound {
			t.Errorf("code: %q", body.Code)
		}
		if body.RequestID == "" {
			t.Error("request id missing from body")
		}
		if w.Header().Get(RequestIDHeader) != body.RequestID {
			t.Errorf("response header %q != body %q",
				w.Header().Get(RequestIDHeader), body.RequestID)
		}
	})

	t.Run("honours client-supplied request id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		req.Header.Set(RequestIDHeader, "client-rid-42")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		var body AppError
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.RequestID != "client-rid-42" {
			t.Errorf("request id: %q", body.RequestID)
		}
	})
}

func TestRecoverMiddleware(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}), RequestIDMiddleware, RecoverMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/explode", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status: %d", w.Code)
	}
	var body AppError
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != CodeInternal {
		t.Errorf("code: %q", body.Code)
	}
	if !strings.Contains(string(w.Body.Bytes()), `"code":"INTERNAL_ERROR"`) {
		t.Errorf("body: %s", w.Body.String())
	}
}

func TestAsAppError(t *testing.T) {
	t.Run("returns same instance", func(t *testing.T) {
		e := ErrEntityNotFound("x", nil)
		if got := AsAppError(e); got != e {
			t.Fatal("should return same pointer")
		}
	})
	t.Run("wraps plain error in Internal", func(t *testing.T) {
		got := AsAppError(errors.New("plain"))
		if got.Code != CodeInternal {
			t.Fatalf("code: %q", got.Code)
		}
		if got.StatusCode != http.StatusInternalServerError {
			t.Fatalf("status: %d", got.StatusCode)
		}
	})
	t.Run("non-nil for nil input", func(t *testing.T) {
		got := AsAppError(nil)
		if got == nil || got.Code != CodeInternal {
			t.Fatalf("got %+v", got)
		}
	})
}

func TestRequestIDFromContext(t *testing.T) {
	ctx := ContextWithRequestID(context.Background(), "rid-9")
	if got := RequestIDFromContext(ctx); got != "rid-9" {
		t.Fatalf("got %q", got)
	}
	if got := RequestIDFromContext(context.Background()); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	if got := RequestIDFromContext(nil); got != "" {
		t.Fatalf("expected empty for nil ctx, got %q", got)
	}
}
