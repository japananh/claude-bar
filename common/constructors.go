package common

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// EntityDetails is the standard `details` payload for entity errors.
// Frontend keys i18n templates off Code and substitutes Entity.
type EntityDetails struct {
	Entity string `json:"entity"`
}

// FieldError describes a single validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// normalizeEntity lowercases the entity for display in messages.
// Capitalisation in error keys is no longer relevant — code is stable.
func normalizeEntity(entity string) string {
	return strings.ToLower(strings.TrimSpace(entity))
}

// ---- Entity helpers -------------------------------------------------------

// ErrEntityNotFound → 404 Not Found.
func ErrEntityNotFound(entity string, root error) *AppError {
	e := normalizeEntity(entity)
	return New(http.StatusNotFound, CodeEntityNotFound,
		fmt.Sprintf("%s not found", e), root).
		WithDetails(EntityDetails{Entity: e})
}

// ErrEntityAlreadyExists → 409 Conflict.
func ErrEntityAlreadyExists(entity string, root error) *AppError {
	e := normalizeEntity(entity)
	return New(http.StatusConflict, CodeEntityAlreadyExists,
		fmt.Sprintf("%s already exists", e), root).
		WithDetails(EntityDetails{Entity: e})
}

// ErrCannotCreateEntity → 422 Unprocessable Entity.
func ErrCannotCreateEntity(entity string, root error) *AppError {
	e := normalizeEntity(entity)
	return New(http.StatusUnprocessableEntity, CodeEntityCreateFailed,
		fmt.Sprintf("Cannot create %s", e), root).
		WithDetails(EntityDetails{Entity: e})
}

// ErrCannotUpdateEntity → 422 Unprocessable Entity.
func ErrCannotUpdateEntity(entity string, root error) *AppError {
	e := normalizeEntity(entity)
	return New(http.StatusUnprocessableEntity, CodeEntityUpdateFailed,
		fmt.Sprintf("Cannot update %s", e), root).
		WithDetails(EntityDetails{Entity: e})
}

// ErrCannotDeleteEntity → 422 Unprocessable Entity.
func ErrCannotDeleteEntity(entity string, root error) *AppError {
	e := normalizeEntity(entity)
	return New(http.StatusUnprocessableEntity, CodeEntityDeleteFailed,
		fmt.Sprintf("Cannot delete %s", e), root).
		WithDetails(EntityDetails{Entity: e})
}

// ErrCannotGetEntity → 422 Unprocessable Entity.
func ErrCannotGetEntity(entity string, root error) *AppError {
	e := normalizeEntity(entity)
	return New(http.StatusUnprocessableEntity, CodeEntityGetFailed,
		fmt.Sprintf("Cannot get %s", e), root).
		WithDetails(EntityDetails{Entity: e})
}

// ErrCannotListEntity → 422 Unprocessable Entity.
func ErrCannotListEntity(entity string, root error) *AppError {
	e := normalizeEntity(entity)
	return New(http.StatusUnprocessableEntity, CodeEntityListFailed,
		fmt.Sprintf("Cannot list %s", e), root).
		WithDetails(EntityDetails{Entity: e})
}

// ---- Auth -----------------------------------------------------------------

// ErrUnauthorized → 401. Use when the caller has no credentials at all.
func ErrUnauthorized(root error) *AppError {
	return New(http.StatusUnauthorized, CodeUnauthorized,
		"Authentication required", root)
}

// ErrInvalidCredentials → 401. Use when credentials were supplied but wrong.
func ErrInvalidCredentials(root error) *AppError {
	return New(http.StatusUnauthorized, CodeInvalidCredentials,
		"Invalid email or password", root)
}

// ErrTokenExpired → 401. Use when a valid token has aged out.
func ErrTokenExpired(root error) *AppError {
	return New(http.StatusUnauthorized, CodeTokenExpired,
		"Session expired, please sign in again", root)
}

// ErrForbidden → 403. Use when caller is authenticated but lacks permission.
func ErrForbidden(root error) *AppError {
	return New(http.StatusForbidden, CodeForbidden,
		"You do not have permission to perform this action", root)
}

// ---- Request / validation -------------------------------------------------

// ErrInvalidRequest → 400. Malformed body, missing required header, etc.
func ErrInvalidRequest(root error) *AppError {
	return New(http.StatusBadRequest, CodeInvalidRequest,
		"Invalid request", root)
}

// ErrValidation → 422. Use when one or more fields fail validation.
// The fields slice is surfaced under details.fields for the frontend.
func ErrValidation(fields []FieldError, root error) *AppError {
	return New(http.StatusUnprocessableEntity, CodeValidationFailed,
		"One or more fields are invalid", root).
		WithDetails(map[string]any{"fields": fields})
}

// ErrWeakPassword → 422.
func ErrWeakPassword(root error) *AppError {
	return New(http.StatusUnprocessableEntity, CodeWeakPassword,
		"Password does not meet the required strength", root)
}

// ErrRateLimited → 429. Optionally surface a retry hint via details.
func ErrRateLimited(retryAfterSeconds int, root error) *AppError {
	return New(http.StatusTooManyRequests, CodeRateLimited,
		"Too many requests, please slow down", root).
		WithDetails(map[string]any{"retry_after_seconds": retryAfterSeconds})
}

// ErrConflict → 409. Generic concurrency / state conflict.
func ErrConflict(message string, root error) *AppError {
	if message == "" {
		message = "Resource state conflict"
	}
	return New(http.StatusConflict, CodeConflict, message, root)
}

// ErrUnsupportedMediaType → 415.
func ErrUnsupportedMediaType(root error) *AppError {
	return New(http.StatusUnsupportedMediaType, CodeUnsupportedMedia,
		"Unsupported media type", root)
}

// ErrPayloadTooLarge → 413.
func ErrPayloadTooLarge(root error) *AppError {
	return New(http.StatusRequestEntityTooLarge, CodePayloadTooLarge,
		"Request payload is too large", root)
}

// ---- Infrastructure -------------------------------------------------------

// ErrDB → 500. The user-facing message is intentionally vague; the real
// SQL error is captured in LogLine() for the server log.
func ErrDB(root error) *AppError {
	return New(http.StatusInternalServerError, CodeDatabaseError,
		"A database error occurred", root)
}

// ErrInternal → 500. Catch-all for unexpected failures.
func ErrInternal(root error) *AppError {
	return New(http.StatusInternalServerError, CodeInternal,
		"Something went wrong", root)
}

// ErrUpstream → 502. Use when a dependency (HTTP API, RPC, queue) fails.
func ErrUpstream(service string, root error) *AppError {
	msg := "Upstream service error"
	if service != "" {
		msg = fmt.Sprintf("Upstream service %q is unavailable", service)
	}
	return New(http.StatusBadGateway, CodeUpstreamError, msg, root).
		WithDetails(map[string]any{"service": service})
}

// ErrTimeout → 504.
func ErrTimeout(operation string, root error) *AppError {
	msg := "Operation timed out"
	if operation != "" {
		msg = fmt.Sprintf("%s timed out", operation)
	}
	return New(http.StatusGatewayTimeout, CodeTimeout, msg, root).
		WithDetails(map[string]any{"operation": operation})
}

// ---- Sentinels ------------------------------------------------------------

// ErrRecordNotFound is the canonical sentinel for repository misses.
// Wrap it via fmt.Errorf("...: %w", common.ErrRecordNotFound) and the
// handler can detect it with errors.Is.
var ErrRecordNotFound = errors.New("record not found")
