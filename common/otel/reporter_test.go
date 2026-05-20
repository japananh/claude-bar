package otel_test

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/soi/backend/pkg/common"
	commonotel "github.com/soi/backend/pkg/common/otel"
)

// newTracerWithRecorder returns a tracer plus a recorder so tests can
// inspect emitted spans synchronously after span.End().
func newTracerWithRecorder(t *testing.T) (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	t.Helper()
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	return tp, rec
}

func TestReporter_5xxSetsErrorStatusAndAttributes(t *testing.T) {
	tp, rec := newTracerWithRecorder(t)

	ctx, span := tp.Tracer("test").Start(context.Background(), "GET /users")
	commonotel.Reporter(ctx, common.ErrDB(errors.New("postgres: timeout")).WithRequestID("rid-1"))
	span.End()

	spans := rec.Ended()
	if len(spans) != 1 {
		t.Fatalf("spans: %d", len(spans))
	}
	s := spans[0]
	if s.Status().Code != otelcodes.Error {
		t.Errorf("status code: got %v want Error", s.Status().Code)
	}
	if s.Status().Description == "" {
		t.Error("status description should hold the log line")
	}

	attrs := attrMap(s.Attributes())
	if got := attrs[commonotel.AttrErrorCode]; got != common.CodeDatabaseError {
		t.Errorf("attr code: %v", got)
	}
	if got := attrs[commonotel.AttrErrorStatus]; got != int64(500) {
		t.Errorf("attr status: %v", got)
	}
	if got := attrs[commonotel.AttrRequestID]; got != "rid-1" {
		t.Errorf("attr request_id: %v", got)
	}
	if len(s.Events()) == 0 {
		t.Error("RecordError should have produced a span event")
	}
}

func TestReporter_4xxAttributesOnlyNoStatusChange(t *testing.T) {
	tp, rec := newTracerWithRecorder(t)

	ctx, span := tp.Tracer("test").Start(context.Background(), "GET /users/x")
	commonotel.Reporter(ctx, common.ErrEntityNotFound("user", nil).WithRequestID("rid-2"))
	span.End()

	s := rec.Ended()[0]
	if s.Status().Code == otelcodes.Error {
		t.Error("4xx must not flip span status to Error")
	}
	attrs := attrMap(s.Attributes())
	if attrs[commonotel.AttrErrorCode] != common.CodeEntityNotFound {
		t.Errorf("attr code: %v", attrs[commonotel.AttrErrorCode])
	}
	if attrs[commonotel.AttrErrorStatus] != int64(404) {
		t.Errorf("attr status: %v", attrs[commonotel.AttrErrorStatus])
	}
	if len(s.Events()) != 0 {
		t.Errorf("4xx must not record error events, got %d", len(s.Events()))
	}
}

func TestReporter_NoSpanInContextIsNoOp(t *testing.T) {
	// No tracer provider, no recording span → must not panic.
	commonotel.Reporter(context.Background(),
		common.ErrInternal(errors.New("x")))
}

func TestReporter_NilAppErrorIsNoOp(t *testing.T) {
	commonotel.Reporter(context.Background(), nil)
}

func TestInstall_WiresReporterIntoCommon(t *testing.T) {
	t.Cleanup(commonotel.Uninstall)
	commonotel.Install()
	if common.Reporter() == nil {
		t.Fatal("Install did not register the reporter")
	}
	commonotel.Uninstall()
	if common.Reporter() != nil {
		t.Fatal("Uninstall did not clear the reporter")
	}
}

func TestReporter_OmitsRequestIDAttrWhenEmpty(t *testing.T) {
	tp, rec := newTracerWithRecorder(t)

	ctx, span := tp.Tracer("test").Start(context.Background(), "op")
	commonotel.Reporter(ctx, common.ErrInvalidRequest(errors.New("bad")))
	span.End()

	attrs := attrMap(rec.Ended()[0].Attributes())
	if _, present := attrs[commonotel.AttrRequestID]; present {
		t.Errorf("request_id attr should be omitted when empty")
	}
}

// attrMap flattens an attribute slice keyed by attribute name into a
// plain map[string]any using the value's native type.
func attrMap(kvs []attribute.KeyValue) map[string]any {
	out := make(map[string]any, len(kvs))
	for _, kv := range kvs {
		out[string(kv.Key)] = kv.Value.AsInterface()
	}
	return out
}
