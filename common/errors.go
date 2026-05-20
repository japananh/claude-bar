package common

// Error codes are stable identifiers used by the frontend for i18n.
// Add new codes here so they are discoverable in one place.
const (
	CodeEntityNotFound      = "ENTITY_NOT_FOUND"
	CodeEntityAlreadyExists = "ENTITY_ALREADY_EXISTS"
	CodeEntityCreateFailed  = "ENTITY_CREATE_FAILED"
	CodeEntityUpdateFailed  = "ENTITY_UPDATE_FAILED"
	CodeEntityDeleteFailed  = "ENTITY_DELETE_FAILED"
	CodeEntityGetFailed     = "ENTITY_GET_FAILED"
	CodeEntityListFailed    = "ENTITY_LIST_FAILED"

	CodeUnauthorized       = "UNAUTHORIZED"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeForbidden          = "FORBIDDEN"
	CodeTokenExpired       = "TOKEN_EXPIRED"

	CodeInvalidRequest    = "INVALID_REQUEST"
	CodeValidationFailed  = "VALIDATION_FAILED"
	CodeWeakPassword      = "WEAK_PASSWORD"
	CodeRateLimited       = "RATE_LIMITED"
	CodeUnsupportedMedia  = "UNSUPPORTED_MEDIA_TYPE"
	CodePayloadTooLarge   = "PAYLOAD_TOO_LARGE"
	CodeConflict          = "CONFLICT"

	CodeDatabaseError = "DATABASE_ERROR"
	CodeInternal      = "INTERNAL_ERROR"
	CodeUpstreamError = "UPSTREAM_ERROR"
	CodeTimeout       = "TIMEOUT"
)

// AppError is the standard error envelope returned by HTTP handlers.
//
// The JSON shape is the contract with the frontend. Server-only fields
// (root, log) are unexported and never serialized.
type AppError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
	RequestID  string `json:"request_id,omitempty"`

	root error
	log  string
}

// New constructs an AppError. If root is non-nil, its Error() is used
// as the log line; otherwise message is logged.
func New(status int, code, message string, root error) *AppError {
	logLine := message
	if root != nil {
		logLine = root.Error()
	}
	return &AppError{
		StatusCode: status,
		Code:       code,
		Message:    message,
		root:       root,
		log:        logLine,
	}
}

// Error implements the error interface. Returns the root cause when
// available, otherwise the user-facing message.
func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.root != nil {
		return e.root.Error()
	}
	return e.Message
}

// Unwrap supports errors.Is / errors.As.
func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.root
}

// LogLine returns the server-side log message (never serialized).
func (e *AppError) LogLine() string {
	if e == nil {
		return ""
	}
	return e.log
}

// WithDetails attaches per-error context to the response body.
// Returns the receiver so calls can be chained.
func (e *AppError) WithDetails(d any) *AppError {
	if e == nil {
		return nil
	}
	e.Details = d
	return e
}

// WithRequestID stamps the request ID on the response body.
func (e *AppError) WithRequestID(id string) *AppError {
	if e == nil {
		return nil
	}
	e.RequestID = id
	return e
}

// WithLog overrides the default log line (which is root.Error() or message).
// Use when you want a richer log line than the user-facing message.
func (e *AppError) WithLog(log string) *AppError {
	if e == nil {
		return nil
	}
	e.log = log
	return e
}

// RetryAfterSeconds extracts the retry hint from details when present.
// Returns 0 if absent or unparseable. WriteJSON consults this to emit
// the standard Retry-After response header.
func (e *AppError) RetryAfterSeconds() int {
	if e == nil {
		return 0
	}
	m, ok := e.Details.(map[string]any)
	if !ok {
		return 0
	}
	v, ok := m["retry_after_seconds"]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		if n > 0 {
			return n
		}
	case int64:
		if n > 0 {
			return int(n)
		}
	case float64:
		if n > 0 {
			return int(n)
		}
	}
	return 0
}
