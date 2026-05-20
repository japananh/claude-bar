// Package otel adapts common.AppError onto OpenTelemetry spans.
//
// Wire it once at startup:
//
//	import (
//	    "github.com/soi/backend/pkg/common"
//	    commonotel "github.com/soi/backend/pkg/common/otel"
//	)
//
//	func main() {
//	    // … set up TracerProvider …
//	    commonotel.Install()
//	    // … start HTTP server …
//	}
//
// After Install, every WriteJSON call decorates the active span with:
//
//	app.error.code   = "ENTITY_NOT_FOUND"
//	app.error.status = 404
//	app.request_id   = "<request id>"
//
// For 5xx responses, the span status is set to Error and the AppError is
// recorded as an event so it surfaces in your APM.
package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/soi/backend/pkg/common"
)

// Attribute keys, exported so tests and tag-search code can refer to them.
const (
	AttrErrorCode   = "app.error.code"
	AttrErrorStatus = "app.error.status"
	AttrRequestID   = "app.request_id"
)

// Reporter is a common.ReporterFunc that records AppErrors onto the
// active span. Pass to common.SetReporter directly, or call Install.
//
// Behaviour:
//   - No-op if no recording span is in the context.
//   - Always attaches code / status / request_id attributes.
//   - Status >= 500 → span.SetStatus(Error, log_line) + span.RecordError(err).
//   - 4xx → attributes only. Client errors do not mark the span as a
//     server fault — that matches OpenTelemetry HTTP semantic conventions.
func Reporter(ctx context.Context, err *common.AppError) {
	if err == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	attrs := make([]attribute.KeyValue, 0, 3)
	attrs = append(attrs,
		attribute.String(AttrErrorCode, err.Code),
		attribute.Int(AttrErrorStatus, err.StatusCode),
	)
	if err.RequestID != "" {
		attrs = append(attrs, attribute.String(AttrRequestID, err.RequestID))
	}
	span.SetAttributes(attrs...)

	if err.StatusCode >= 500 {
		span.SetStatus(codes.Error, err.LogLine())
		span.RecordError(err)
	}
}

// Install wires Reporter into the common package's global reporter slot.
// Calling Install twice is safe; the latest call wins.
func Install() {
	common.SetReporter(Reporter)
}

// Uninstall removes the reporter. Useful in tests.
func Uninstall() {
	common.SetReporter(nil)
}
