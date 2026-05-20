// Package common provides the standard error envelope and HTTP helpers
// shared between the backend and the frontend.
//
// # Wire contract
//
// Every error response has the same JSON shape:
//
//	{
//	  "code":       "ENTITY_NOT_FOUND",
//	  "message":    "user not found",
//	  "details":    { "entity": "user" },
//	  "request_id": "c1f2a3b4..."
//	}
//
//   - code: stable machine identifier in SCREAMING_SNAKE_CASE. Frontend
//     keys i18n tables off this. Never changes per entity.
//   - message: English fallback. Frontend uses it only if it doesn't
//     have a translation for code.
//   - details: per-error context (entity name, validation fields, etc.).
//   - request_id: propagated from RequestIDMiddleware for support
//     correlation between client bug reports and server logs.
//
// # Server-only fields
//
// The wrapped root error and the log line are NEVER serialized. Use
// LogLine() to write them to your structured logger.
//
// # Usage
//
//	// In a handler:
//	user, err := repo.Get(ctx, id)
//	if errors.Is(err, common.ErrRecordNotFound) {
//	    common.WriteJSON(w, r, common.ErrEntityNotFound("user", err))
//	    return
//	}
//	if err != nil {
//	    common.WriteJSON(w, r, common.ErrDB(err))
//	    return
//	}
package common
