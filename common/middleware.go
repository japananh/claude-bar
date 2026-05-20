package common

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"runtime/debug"
)

// ctxKey is unexported so external packages cannot collide on the same key.
type ctxKey struct{ name string }

var requestIDCtxKey = ctxKey{name: "request_id"}

// RequestIDHeader is the canonical header name. Read on the way in,
// echoed on the way out so frontends and proxies can correlate.
const RequestIDHeader = "X-Request-ID"

// RequestIDMiddleware ensures every request has a request ID. If the
// client provides one in X-Request-ID, it is honoured; otherwise a new
// 128-bit hex token is generated. The ID is attached to the request
// context (use RequestIDFromContext) and echoed back via the response
// header.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(RequestIDHeader)
		if rid == "" {
			rid = newRequestID()
		}
		ctx := context.WithValue(r.Context(), requestIDCtxKey, rid)
		w.Header().Set(RequestIDHeader, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext returns the request ID stamped by RequestIDMiddleware,
// or an empty string if none.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	rid, _ := ctx.Value(requestIDCtxKey).(string)
	return rid
}

// ContextWithRequestID is a test helper that injects a request ID into ctx.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDCtxKey, id)
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// rand.Read should never fail on supported platforms; fall back
		// to a deterministic-but-unique-enough sentinel rather than panic.
		return "rid-fallback"
	}
	return hex.EncodeToString(b[:])
}

// RecoverMiddleware converts panics into 500 responses with a stable
// JSON shape and a stack trace in the server log.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			err := fmt.Errorf("panic: %v", rec)
			WriteJSON(w, r, ErrInternal(err).WithLog(
				fmt.Sprintf("panic: %v\n%s", rec, debug.Stack()),
			))
		}()
		next.ServeHTTP(w, r)
	})
}

// Chain composes middleware in the order given. The first middleware
// in the list is the outermost wrapper. Example:
//
//	handler := common.Chain(mux,
//	    common.RequestIDMiddleware,
//	    common.RecoverMiddleware,
//	)
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
