/**
 * i18n tables keyed by stable error code.
 *
 * The table never changes when the backend adds a new entity — `entity`
 * comes from `details.entity` and is substituted at render time.
 */

import {
  AppError,
  ErrorCode,
  type EntityDetails,
  type ValidationDetails,
} from "./errors";

export type Locale = "en" | "vi";

type Renderer = (details?: Record<string, unknown>) => string;

const entity = (details?: Record<string, unknown>): string =>
  (details as EntityDetails | undefined)?.entity ?? "resource";

export const MESSAGES_EN: Record<string, Renderer> = {
  [ErrorCode.EntityNotFound]: (d) => `${entity(d)} not found`,
  [ErrorCode.EntityAlreadyExists]: (d) => `${entity(d)} already exists`,
  [ErrorCode.EntityCreateFailed]: (d) => `Cannot create ${entity(d)}`,
  [ErrorCode.EntityUpdateFailed]: (d) => `Cannot update ${entity(d)}`,
  [ErrorCode.EntityDeleteFailed]: (d) => `Cannot delete ${entity(d)}`,
  [ErrorCode.EntityGetFailed]: (d) => `Cannot load ${entity(d)}`,
  [ErrorCode.EntityListFailed]: (d) => `Cannot list ${entity(d)}`,

  [ErrorCode.Unauthorized]: () => "Please sign in to continue",
  [ErrorCode.InvalidCredentials]: () => "Invalid email or password",
  [ErrorCode.Forbidden]: () =>
    "You do not have permission to perform this action",
  [ErrorCode.TokenExpired]: () => "Your session has expired, please sign in again",

  [ErrorCode.InvalidRequest]: () => "Invalid request",
  [ErrorCode.ValidationFailed]: () => "Please check the highlighted fields",
  [ErrorCode.WeakPassword]: () => "Password is not strong enough",
  [ErrorCode.RateLimited]: (d) => {
    const seconds = (d as { retry_after_seconds?: number } | undefined)
      ?.retry_after_seconds;
    return seconds
      ? `Too many requests, please retry in ${seconds}s`
      : "Too many requests, please slow down";
  },
  [ErrorCode.Conflict]: () => "This resource has changed, please refresh and try again",
  [ErrorCode.UnsupportedMedia]: () => "Unsupported file type",
  [ErrorCode.PayloadTooLarge]: () => "The file is too large",

  [ErrorCode.DatabaseError]: () => "We are having trouble right now, please try again",
  [ErrorCode.Internal]: () => "Something went wrong, please try again",
  [ErrorCode.UpstreamError]: () => "An upstream service is unavailable",
  [ErrorCode.Timeout]: () => "The request took too long, please try again",

  [ErrorCode.NetworkError]: () => "Network error — check your connection",
  [ErrorCode.ParseError]: () => "Unexpected response from the server",
};

export const MESSAGES_VI: Record<string, Renderer> = {
  [ErrorCode.EntityNotFound]: (d) => `Không tìm thấy ${entity(d)}`,
  [ErrorCode.EntityAlreadyExists]: (d) => `${entity(d)} đã tồn tại`,
  [ErrorCode.EntityCreateFailed]: (d) => `Không thể tạo ${entity(d)}`,
  [ErrorCode.EntityUpdateFailed]: (d) => `Không thể cập nhật ${entity(d)}`,
  [ErrorCode.EntityDeleteFailed]: (d) => `Không thể xoá ${entity(d)}`,
  [ErrorCode.EntityGetFailed]: (d) => `Không thể tải ${entity(d)}`,
  [ErrorCode.EntityListFailed]: (d) => `Không thể tải danh sách ${entity(d)}`,

  [ErrorCode.Unauthorized]: () => "Vui lòng đăng nhập để tiếp tục",
  [ErrorCode.InvalidCredentials]: () => "Email hoặc mật khẩu không đúng",
  [ErrorCode.Forbidden]: () => "Bạn không có quyền thực hiện thao tác này",
  [ErrorCode.TokenExpired]: () => "Phiên đăng nhập đã hết hạn, vui lòng đăng nhập lại",

  [ErrorCode.InvalidRequest]: () => "Yêu cầu không hợp lệ",
  [ErrorCode.ValidationFailed]: () => "Vui lòng kiểm tra lại các trường được đánh dấu",
  [ErrorCode.WeakPassword]: () => "Mật khẩu chưa đủ mạnh",
  [ErrorCode.RateLimited]: (d) => {
    const seconds = (d as { retry_after_seconds?: number } | undefined)
      ?.retry_after_seconds;
    return seconds
      ? `Quá nhiều yêu cầu, vui lòng thử lại sau ${seconds}s`
      : "Quá nhiều yêu cầu, vui lòng thử lại sau";
  },
  [ErrorCode.Conflict]: () =>
    "Dữ liệu đã thay đổi, vui lòng tải lại và thử lại",
  [ErrorCode.UnsupportedMedia]: () => "Định dạng tệp không được hỗ trợ",
  [ErrorCode.PayloadTooLarge]: () => "Tệp tải lên quá lớn",

  [ErrorCode.DatabaseError]: () => "Hệ thống đang gặp sự cố, vui lòng thử lại",
  [ErrorCode.Internal]: () => "Đã có lỗi xảy ra, vui lòng thử lại",
  [ErrorCode.UpstreamError]: () => "Dịch vụ bên ngoài đang không khả dụng",
  [ErrorCode.Timeout]: () => "Yêu cầu mất quá nhiều thời gian, vui lòng thử lại",

  [ErrorCode.NetworkError]: () => "Lỗi kết nối — kiểm tra mạng của bạn",
  [ErrorCode.ParseError]: () => "Phản hồi từ máy chủ không hợp lệ",
};

const TABLES: Record<Locale, Record<string, Renderer>> = {
  en: MESSAGES_EN,
  vi: MESSAGES_VI,
};

/**
 * Localise an AppError to a user-facing string.
 *
 * Falls back to the backend's English `message` if no translation exists
 * for the code in the requested locale.
 */
export function localize(err: AppError, locale: Locale = "en"): string {
  const table = TABLES[locale] ?? MESSAGES_EN;
  const renderer = table[err.code];
  if (renderer) return renderer(err.details);
  return err.message;
}

/**
 * Localise a per-field validation error. Falls back to the server message
 * if no client override is configured.
 */
export function localizeFieldError(
  err: AppError,
  locale: Locale = "en",
): Record<string, string> {
  const fields = (err.details as ValidationDetails | undefined)?.fields;
  if (!Array.isArray(fields)) return {};
  const out: Record<string, string> = {};
  for (const f of fields) {
    out[f.field] = f.message ?? localize(err, locale);
  }
  return out;
}
