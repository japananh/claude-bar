/**
 * Thin fetch wrapper that converts any non-2xx response into an AppError.
 *
 * Design goals:
 *   - Throw AppError (never a raw Response) so callers use one catch path.
 *   - Preserve request_id from the server for support correlation.
 *   - Add the X-Request-ID header on the way out so the backend's log
 *     and the browser dev tools agree on a single ID.
 */

import {
  AppError,
  ErrorCode,
  type AppErrorPayload,
} from "./errors";

export interface HttpClientOptions {
  /** Base URL prepended to every request. e.g. "https://api.example.com" */
  baseUrl?: string;
  /** Default headers merged into every request. */
  defaultHeaders?: Record<string, string>;
  /** Per-request timeout in ms (0 = no timeout). Default: 15_000. */
  timeoutMs?: number;
  /** Generator for X-Request-ID. Defaults to crypto.randomUUID. */
  requestIdGenerator?: () => string;
  /** Hook called whenever an AppError is about to be thrown. */
  onError?: (err: AppError) => void;
}

export interface RequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown;
  /** Skip baseUrl prefix (use when calling absolute URLs). */
  absolute?: boolean;
  /** Override the client-level timeout for this call. */
  timeoutMs?: number;
}

const REQUEST_ID_HEADER = "X-Request-ID";

function defaultRequestId(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  // Fallback for environments without crypto.randomUUID.
  return `rid-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export class HttpClient {
  private readonly baseUrl: string;
  private readonly defaultHeaders: Record<string, string>;
  private readonly timeoutMs: number;
  private readonly genId: () => string;
  private readonly onError?: (err: AppError) => void;

  constructor(opts: HttpClientOptions = {}) {
    this.baseUrl = (opts.baseUrl ?? "").replace(/\/+$/, "");
    this.defaultHeaders = opts.defaultHeaders ?? {};
    this.timeoutMs = opts.timeoutMs ?? 15_000;
    this.genId = opts.requestIdGenerator ?? defaultRequestId;
    this.onError = opts.onError;
  }

  get<T>(path: string, opts?: RequestOptions): Promise<T> {
    return this.request<T>("GET", path, opts);
  }
  post<T>(path: string, body?: unknown, opts?: RequestOptions): Promise<T> {
    return this.request<T>("POST", path, { ...opts, body });
  }
  put<T>(path: string, body?: unknown, opts?: RequestOptions): Promise<T> {
    return this.request<T>("PUT", path, { ...opts, body });
  }
  patch<T>(path: string, body?: unknown, opts?: RequestOptions): Promise<T> {
    return this.request<T>("PATCH", path, { ...opts, body });
  }
  delete<T>(path: string, opts?: RequestOptions): Promise<T> {
    return this.request<T>("DELETE", path, opts);
  }

  async request<T>(
    method: string,
    path: string,
    opts: RequestOptions = {},
  ): Promise<T> {
    const url = opts.absolute ? path : `${this.baseUrl}${path}`;
    const requestId = this.genId();

    const headers = new Headers(this.defaultHeaders);
    if (opts.headers) {
      new Headers(opts.headers as HeadersInit).forEach((v, k) => headers.set(k, v));
    }
    headers.set(REQUEST_ID_HEADER, requestId);

    let body: BodyInit | undefined;
    if (opts.body !== undefined && opts.body !== null) {
      if (opts.body instanceof FormData || opts.body instanceof Blob) {
        body = opts.body as BodyInit;
      } else if (typeof opts.body === "string") {
        body = opts.body;
      } else {
        body = JSON.stringify(opts.body);
        if (!headers.has("Content-Type")) {
          headers.set("Content-Type", "application/json; charset=utf-8");
        }
      }
    }

    const timeoutMs = opts.timeoutMs ?? this.timeoutMs;
    const controller = new AbortController();
    const timer =
      timeoutMs > 0 ? setTimeout(() => controller.abort(), timeoutMs) : null;

    let res: Response;
    try {
      res = await fetch(url, {
        method,
        headers,
        body,
        signal: controller.signal,
        credentials: opts.credentials,
        mode: opts.mode,
        cache: opts.cache,
        redirect: opts.redirect,
        referrer: opts.referrer,
        referrerPolicy: opts.referrerPolicy,
        integrity: opts.integrity,
        keepalive: opts.keepalive,
      });
    } catch (cause) {
      if (timer) clearTimeout(timer);
      const err =
        controller.signal.aborted
          ? new AppError(ErrorCode.Timeout, "Request timed out", 0)
          : AppError.networkError(cause);
      this.onError?.(err);
      throw err;
    }
    if (timer) clearTimeout(timer);

    if (res.ok) {
      // 204 No Content
      if (res.status === 204) return undefined as T;
      const contentType = res.headers.get("content-type") ?? "";
      if (contentType.includes("application/json")) {
        return (await res.json()) as T;
      }
      // Caller wants a non-JSON body — return as-is.
      return (await res.text()) as unknown as T;
    }

    const err = await this.errorFromResponse(res);
    this.onError?.(err);
    throw err;
  }

  private async errorFromResponse(res: Response): Promise<AppError> {
    const raw = await res.text();
    let payload: AppErrorPayload | null = null;
    try {
      payload = raw ? (JSON.parse(raw) as AppErrorPayload) : null;
    } catch {
      // body wasn't JSON
    }
    if (payload && typeof payload.code === "string") {
      // server gave us a proper envelope
      if (!payload.request_id) {
        const rid = res.headers.get(REQUEST_ID_HEADER);
        if (rid) payload.request_id = rid;
      }
      return AppError.fromPayload(res.status, payload);
    }
    return AppError.parseError(res.status, raw);
  }
}
