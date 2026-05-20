# common — Standard error envelope (Go + TypeScript)

A single error contract shared between the Go backend and the TypeScript
frontend. One stable JSON shape, one set of error codes, one i18n table.

## Wire contract

Every error response from the backend has this shape:

```json
{
  "code":       "ENTITY_NOT_FOUND",
  "message":    "user not found",
  "details":    { "entity": "user" },
  "request_id": "c1f2a3b4d5e6..."
}
```

| Field        | Stability                          | Purpose                                        |
| ------------ | ---------------------------------- | ---------------------------------------------- |
| `code`       | **Stable** (SCREAMING_SNAKE_CASE)  | Frontend keys i18n + business logic off this   |
| `message`    | English fallback only              | Shown if locale has no translation             |
| `details`    | Per-code shape                     | Entity name, validation fields, retry hints…   |
| `request_id` | Stamped by middleware              | Correlate frontend bug reports with server log |

**Never serialised:** the wrapped root error and the server log line.
The frontend never sees raw SQL errors or stack traces.

## HTTP status map

| Constructor                  | Status | Code                       |
| ---------------------------- | -----: | -------------------------- |
| `ErrEntityNotFound`          |    404 | `ENTITY_NOT_FOUND`         |
| `ErrEntityAlreadyExists`     |    409 | `ENTITY_ALREADY_EXISTS`    |
| `ErrCannotCreateEntity`      |    422 | `ENTITY_CREATE_FAILED`     |
| `ErrCannotUpdateEntity`      |    422 | `ENTITY_UPDATE_FAILED`     |
| `ErrCannotDeleteEntity`      |    422 | `ENTITY_DELETE_FAILED`     |
| `ErrCannotGetEntity`         |    422 | `ENTITY_GET_FAILED`        |
| `ErrCannotListEntity`        |    422 | `ENTITY_LIST_FAILED`       |
| `ErrUnauthorized`            |    401 | `UNAUTHORIZED`             |
| `ErrInvalidCredentials`      |    401 | `INVALID_CREDENTIALS`      |
| `ErrTokenExpired`            |    401 | `TOKEN_EXPIRED`            |
| `ErrForbidden`               |    403 | `FORBIDDEN`                |
| `ErrInvalidRequest`          |    400 | `INVALID_REQUEST`          |
| `ErrValidation`              |    422 | `VALIDATION_FAILED`        |
| `ErrWeakPassword`            |    422 | `WEAK_PASSWORD`            |
| `ErrRateLimited`             |    429 | `RATE_LIMITED`             |
| `ErrConflict`                |    409 | `CONFLICT`                 |
| `ErrUnsupportedMediaType`    |    415 | `UNSUPPORTED_MEDIA_TYPE`   |
| `ErrPayloadTooLarge`         |    413 | `PAYLOAD_TOO_LARGE`        |
| `ErrDB`                      |    500 | `DATABASE_ERROR`           |
| `ErrInternal`                |    500 | `INTERNAL_ERROR`           |
| `ErrUpstream`                |    502 | `UPSTREAM_ERROR`           |
| `ErrTimeout`                 |    504 | `TIMEOUT`                  |

## Layout

```
common/                                core module (stdlib only)
├── go.mod                             github.com/soi/backend/pkg/common
├── errors.go      constructors.go     http.go    middleware.go
├── reporter.go    *_test.go
└── otel/                              opt-in adapter (separate module)
    ├── go.mod                         github.com/soi/backend/pkg/common/otel
    └── reporter.go                    OpenTelemetry span attribution
```

The core module pulls **zero** external dependencies. The OTel adapter
sits behind its own go.mod so consumers who don't use OTel never see it
in their dependency graph.

## Backend usage (Go)

```go
import "github.com/soi/backend/pkg/common"

func getUser(w http.ResponseWriter, r *http.Request) {
    user, err := repo.Get(r.Context(), mux.Vars(r)["id"])
    if errors.Is(err, common.ErrRecordNotFound) {
        common.WriteJSON(w, r, common.ErrEntityNotFound("user", err))
        return
    }
    if err != nil {
        common.WriteJSON(w, r, common.ErrDB(err))
        return
    }
    common.WriteSuccess(w, http.StatusOK, user)
}

// In main():
mux := http.NewServeMux()
mux.HandleFunc("/users/{id}", getUser)

handler := common.Chain(mux,
    common.RequestIDMiddleware,
    common.RecoverMiddleware,
)
http.ListenAndServe(":8080", handler)
```

### Validation

```go
return common.ErrValidation([]common.FieldError{
    {Field: "email", Message: "must be a valid email"},
    {Field: "password", Message: "must be at least 8 characters", Code: "WEAK_PASSWORD"},
}, nil)
```

### Sentinels with `errors.Is`

```go
if errors.Is(err, common.ErrRecordNotFound) {
    // ...
}
```

### Rate limiting + `Retry-After`

`ErrRateLimited(seconds, root)` carries the retry hint in
`details.retry_after_seconds`. `WriteJSON` automatically emits the
standard `Retry-After` response header so HTTP clients (browsers,
fetch, axios, ts client) honour it without custom code.

```go
common.WriteJSON(w, r, common.ErrRateLimited(30, nil))
// → HTTP 429, header:  Retry-After: 30
// → body:   {"code":"RATE_LIMITED","details":{"retry_after_seconds":30},...}
```

### OpenTelemetry integration

The OTel adapter is shipped as a separate sub-module so it adds zero
dependencies to consumers who don't use it.

```go
import (
    "github.com/soi/backend/pkg/common"
    commonotel "github.com/soi/backend/pkg/common/otel"
)

func main() {
    // … set up TracerProvider …
    commonotel.Install()   // one-liner, idempotent
    // … start HTTP server …
}
```

After install, every `WriteJSON` call decorates the active span:

| Attribute          | Value example         |
|--------------------|-----------------------|
| `app.error.code`   | `ENTITY_NOT_FOUND`    |
| `app.error.status` | `404`                 |
| `app.request_id`   | `c1f2a3b4d5e6...`     |

5xx responses additionally set span status to `Error` and record the
error as a span event. 4xx responses attach attributes only — client
errors should not be reported as server faults (matches OTel HTTP
semantic conventions).

### Custom error reporter (Sentry, Datadog, etc.)

The reporter hook is plain `func(ctx, *AppError)` — wire whatever you use:

```go
common.SetReporter(func(ctx context.Context, err *common.AppError) {
    if err.StatusCode < 500 {
        return // skip 4xx
    }
    sentry.CaptureException(err)
})
```

Reporters run synchronously after the response is sent. A panic inside a
reporter is recovered and swallowed so it cannot break the response.

## Frontend usage (TypeScript)

```ts
import { HttpClient, AppError, localize } from "@yourorg/common";

const api = new HttpClient({
  baseUrl: "/api",
  onError: (err) => console.warn(err.code, err.requestId),
});

try {
  const user = await api.get<User>(`/users/${id}`);
} catch (err) {
  if (err instanceof AppError) {
    if (err.requiresAuth())   router.push("/sign-in");
    else if (err.isNotFound()) toast.info(localize(err, "vi"));
    else if (err.validationFields()) {
      form.setErrors(err.validationFields()!);
    } else {
      toast.error(localize(err, "vi"));
    }
  }
}
```

### React Query example

```ts
const { data, error } = useQuery({
  queryKey: ["user", id],
  queryFn: () => api.get<User>(`/users/${id}`),
});

if (error instanceof AppError && error.isNotFound()) {
  return <NotFound entity="user" />;
}
```

## Adding a new error code

1. Add the constant in `errors.go` (`const Code... = "FOO_BAR"`).
2. Add a constructor in `constructors.go` if you want a typed helper.
3. Mirror the constant in `ts/errors.ts` (`ErrorCode.FooBar`).
4. Add the renderer in `ts/i18n.ts` for both `MESSAGES_EN` and `MESSAGES_VI`.
5. Add a row to the HTTP status map in this README.

## Why not RFC 7807 (Problem Details)?

RFC 7807 is fine for public APIs, but its `type` URI and `title`/`detail`
split is awkward for i18n. We use a single `code` (stable, machine-friendly)
plus `details` (typed per code). If you later need to expose a public API,
both shapes can coexist: add a `type` field aliased to a doc URL.

## Testing

Core module (stdlib only):

```bash
cd common && go test ./...
```

OTel adapter (separate module, requires OTel deps):

```bash
cd common/otel && go test ./...
```

Frontend (strict tsc):

```bash
cd common/ts && tsc --noEmit --strict --target ES2020 \
    --moduleResolution node --module ESNext --lib ES2020,DOM \
    errors.ts i18n.ts http-client.ts index.ts
```

## Open questions

- Do you want a separate `audit_id` distinct from `request_id` (e.g. one
  per logical transaction spanning multiple HTTP calls)? Not added yet.
