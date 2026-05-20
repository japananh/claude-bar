package common

import (
	"context"
	"sync/atomic"
)

// ReporterFunc is invoked by WriteJSON after the response is written
// so the error can be propagated to an external system (OpenTelemetry,
// Sentry, Datadog, …).
//
// Implementations must be cheap and must not block. Offload heavy work
// to a goroutine. The function may be called concurrently.
type ReporterFunc func(ctx context.Context, err *AppError)

// reporterHolder uses an atomic pointer so SetReporter is safe under
// concurrent reads without a mutex on the hot path.
var reporterHolder atomic.Pointer[ReporterFunc]

// SetReporter installs the global reporter. Pass nil to remove the hook.
// Safe to call concurrently from any goroutine.
func SetReporter(r ReporterFunc) {
	if r == nil {
		reporterHolder.Store(nil)
		return
	}
	reporterHolder.Store(&r)
}

// Reporter returns the currently installed reporter, or nil if none.
// Exposed mainly for tests.
func Reporter() ReporterFunc {
	p := reporterHolder.Load()
	if p == nil {
		return nil
	}
	return *p
}

// invokeReporter calls the reporter if installed. Recovers from panics
// so a misbehaving reporter cannot poison the response path.
func invokeReporter(ctx context.Context, err *AppError) {
	p := reporterHolder.Load()
	if p == nil || *p == nil || err == nil {
		return
	}
	defer func() { _ = recover() }()
	(*p)(ctx, err)
}
