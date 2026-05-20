/**
 * Standard error envelope shared with the Go backend.
 *
 * Wire shape (must match `common.AppError` JSON):
 * {
 *   "code":       "ENTITY_NOT_FOUND",
 *   "message":    "user not found",
 *   "details":    { "entity": "user" },
 *   "request_id": "c1f2a3b4..."
 * }
 */

export const ErrorCode = {
  // Entity
  EntityNotFound: "ENTITY_NOT_FOUND",
  EntityAlreadyExists: "ENTITY_ALREADY_EXISTS",
  EntityCreateFailed: "ENTITY_CREATE_FAILED",
  EntityUpdateFailed: "ENTITY_UPDATE_FAILED",
  EntityDeleteFailed: "ENTITY_DELETE_FAILED",
  EntityGetFailed: "ENTITY_GET_FAILED",
  EntityListFailed: "ENTITY_LIST_FAILED",

  // Auth
  Unauthorized: "UNAUTHORIZED",
  InvalidCredentials: "INVALID_CREDENTIALS",
  Forbidden: "FORBIDDEN",
  TokenExpired: "TOKEN_EXPIRED",

  // Request / validation
  InvalidRequest: "INVALID_REQUEST",
  ValidationFailed: "VALIDATION_FAILED",
  WeakPassword: "WEAK_PASSWORD",
  RateLimited: "RATE_LIMITED",
  Conflict: "CONFLICT",
  UnsupportedMedia: "UNSUPPORTED_MEDIA_TYPE",
  PayloadTooLarge: "PAYLOAD_TOO_LARGE",

  // Infrastructure
  DatabaseError: "DATABASE_ERROR",
  Internal: "INTERNAL_ERROR",
  UpstreamError: "UPSTREAM_ERROR",
  Timeout: "TIMEOUT",

  // Client-only
  NetworkError: "NETWORK_ERROR",
  ParseError: "PARSE_ERROR",
} as const;

export type ErrorCode = (typeof ErrorCode)[keyof typeof ErrorCode];

/** JSON payload as returned by the backend. */
export interface AppErrorPayload {
  code: ErrorCode | string;
  message: string;
  details?: Record<string, unknown>;
  request_id?: string;
}

/** Optional details shape for entity errors. Frontend uses entity for i18n. */
export interface EntityDetails {
  entity: string;
}

/** Optional details shape for validation errors. */
export interface ValidationDetails {
  fields: FieldError[];
}

export interface FieldError {
  field: string;
  message: string;
  code?: string;
}

/**
 * AppError is the canonical error thrown by the API client.
 *
 * Use `error instanceof AppError` for narrowing. Use `error.code` for
 * i18n / business logic — never key off `error.message`, which is the
 * English fallback only.
 */
export class AppError extends Error {
  readonly code: ErrorCode | string;
  readonly status: number;
  readonly details?: Record<string, unknown>;
  readonly requestId?: string;

  constructor(
    code: ErrorCode | string,
    message: string,
    status: number,
    details?: Record<string, unknown>,
    requestId?: string,
  ) {
    super(message);
    this.name = "AppError";
    this.code = code;
    this.status = status;
    this.details = details;
    this.requestId = requestId;
    // Restore prototype chain (needed for `instanceof` in compiled ES5).
    Object.setPrototypeOf(this, AppError.prototype);
  }

  /** Build an AppError from a fetch Response body. */
  static fromPayload(status: number, body: AppErrorPayload): AppError {
    return new AppError(
      body.code,
      body.message,
      status,
      body.details,
      body.request_id,
    );
  }

  /** Build an AppError for a network failure (fetch threw). */
  static networkError(cause: unknown): AppError {
    const message = cause instanceof Error ? cause.message : String(cause);
    return new AppError(ErrorCode.NetworkError, message, 0);
  }

  /** Build an AppError for a body parse failure (non-JSON or unexpected shape). */
  static parseError(status: number, raw: string): AppError {
    return new AppError(
      ErrorCode.ParseError,
      `Unexpected response (status ${status})`,
      status,
      { raw },
    );
  }

  /** Narrow into a specific code. Useful in catch blocks. */
  is(code: ErrorCode | string): boolean {
    return this.code === code;
  }

  /** Convenience: did this fail because of a missing entity? */
  isNotFound(): boolean {
    return this.code === ErrorCode.EntityNotFound || this.status === 404;
  }

  /** Convenience: requires the user to (re)authenticate? */
  requiresAuth(): boolean {
    return (
      this.code === ErrorCode.Unauthorized ||
      this.code === ErrorCode.TokenExpired ||
      this.status === 401
    );
  }

  /** Convenience: forbidden (authenticated but not allowed)? */
  isForbidden(): boolean {
    return this.code === ErrorCode.Forbidden || this.status === 403;
  }

  /** Convenience: was this a validation error? Returns the fields if so. */
  validationFields(): FieldError[] | null {
    if (this.code !== ErrorCode.ValidationFailed) return null;
    const fields = (this.details as ValidationDetails | undefined)?.fields;
    return Array.isArray(fields) ? fields : [];
  }
}

/**
 * Type guard. Cheaper than `instanceof` if the error crossed a structured-clone
 * boundary (postMessage, web worker) where the prototype was stripped.
 */
export function isAppError(err: unknown): err is AppError {
  if (err instanceof AppError) return true;
  if (typeof err !== "object" || err === null) return false;
  const e = err as Record<string, unknown>;
  return typeof e.code === "string" && typeof e.message === "string";
}
