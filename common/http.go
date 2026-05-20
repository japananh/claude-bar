package common

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
)

// WriteJSON writes err as a JSON response with the appropriate HTTP status.
// Any non-AppError is wrapped in ErrInternal so the response shape is stable.
//
// The request ID (set by RequestIDMiddleware) is stamped on the body and
// the server-side log line is emitted at error level.
//
// Caller convention: WriteJSON terminates the response. Do not write again.
func WriteJSON(w http.ResponseWriter, r *http.Request, err error) {
	appErr := AsAppError(err)
	ctx := r.Context()
	if rid := RequestIDFromContext(ctx); rid != "" && appErr.RequestID == "" {
		appErr.RequestID = rid
	}

	slog.ErrorContext(ctx, "http error",
		slog.String("code", appErr.Code),
		slog.Int("status", appErr.StatusCode),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("request_id", appErr.RequestID),
		slog.String("log", appErr.LogLine()),
	)
	invokeReporter(ctx, appErr)

	if secs := appErr.RetryAfterSeconds(); secs > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(secs))
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(appErr.StatusCode)
	_ = json.NewEncoder(w).Encode(appErr)
}

// AsAppError extracts an *AppError from err, wrapping it in ErrInternal
// if it is not already one. Never returns nil for a non-nil input.
func AsAppError(err error) *AppError {
	if err == nil {
		return ErrInternal(errors.New("nil error passed to AsAppError"))
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return ErrInternal(err)
}

// WriteSuccess writes any value as a JSON 200 response. Provided for
// symmetry with WriteJSON so handlers stay uniform.
func WriteSuccess(w http.ResponseWriter, status int, body any) {
	if status == 0 {
		status = http.StatusOK
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}
