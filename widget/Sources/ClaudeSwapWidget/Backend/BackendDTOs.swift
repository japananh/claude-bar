import Foundation

/// Mirrors backend/internal/domain/account.go Account.
struct AccountDTO: Codable, Hashable {
    let number: Int
    let email: String
    let organizationName: String?
    let organizationUuid: String?
    let nickname: String?
    let createdAt: Date

    var displayName: String {
        if let n = nickname, !n.isEmpty { return n }
        return email
    }

    /// True when the org name is the auto-generated personal placeholder
    /// (e.g. "alice@example.com's Organization") and worth hiding.
    var hasMeaningfulOrg: Bool {
        guard let org = organizationName, !org.isEmpty else { return false }
        return org != "\(email)'s Organization"
    }

    /// One-letter initial for the avatar circle.
    var initial: String {
        let src = displayName.trimmingCharacters(in: .whitespaces)
        return src.isEmpty ? "?" : String(src.prefix(1)).uppercased()
    }
}

/// Mirrors backend/internal/domain/usage.go Window.
///
/// `utilizationPct` is already a percentage in [0, 100] (the Anthropic
/// usage API returns "utilization" as a percent, not a fraction).
struct UsageWindowDTO: Codable, Hashable {
    let utilizationPct: Double
    let resetsAt: Date

    var percentInt: Int { Int(utilizationPct.rounded()) }
    var fractionForBar: Double { max(0, min(1, utilizationPct / 100)) }

    func secondsUntilReset(now: Date = Date()) -> Int {
        max(0, Int(resetsAt.timeIntervalSince(now)))
    }

    /// Consistent reset countdown: "12m", "3h 49m", "2d 8h".
    func resetLabel(now: Date = Date()) -> String {
        let secs = secondsUntilReset(now: now)
        if secs == 0 { return "now" }
        let h = secs / 3600
        let m = (secs % 3600) / 60
        if h >= 24 {
            let d = h / 24
            let rh = h % 24
            return rh == 0 ? "\(d)d" : "\(d)d \(rh)h"
        }
        if h >= 1 { return "\(h)h \(m)m" }
        return "\(max(1, m))m"
    }
}

/// Mirrors backend/internal/domain/usage.go Usage.
struct UsageDTO: Codable, Hashable {
    let fiveHour: UsageWindowDTO?
    let sevenDay: UsageWindowDTO?
    let fetchedAt: Date
}

/// Mirrors backend/internal/usecase/list_accounts.go AccountView.
struct AccountViewDTO: Codable, Hashable, Identifiable {
    let account: AccountDTO
    let isActive: Bool
    let usage: UsageDTO?
    let error: String?

    var id: Int { account.number }
}

/// Mirrors backend ListAccountsResult.
struct ListAccountsDTO: Codable, Hashable {
    let accounts: [AccountViewDTO]
    let activeAccountNumber: Int

    var active: AccountViewDTO? { accounts.first(where: { $0.isActive }) }
}

/// Mirrors backend/internal/domain/session.go SessionReport.
struct SessionReportDTO: Codable, Hashable {
    let total: Int
    let busyOrWaiting: Int
    let interactiveOnly: Int
    let safeToSwap: Bool
}

/// Mirrors backend AddAccountResult.
struct AddAccountDTO: Codable, Hashable {
    let account: AccountDTO
    let wasDuplicate: Bool
    let duplicateOfNum: Int?
}

/// Mirrors backend/internal/domain/verification.go CheckResult.
struct CheckResultDTO: Codable, Hashable, Identifiable {
    let name: String
    let passed: Bool
    var skipped: Bool? = nil
    var detail: String? = nil

    var id: String { name }
    var label: String {
        switch name {
        case "credentials_present": return "Credentials present"
        case "credentials_valid":   return "Credentials valid"
        case "token_refresh":       return "Token refresh"
        case "usage_reachable":     return "Usage API reachable"
        default:                    return name.replacingOccurrences(of: "_", with: " ").capitalized
        }
    }
}

/// Mirrors AccountVerification.
struct AccountVerificationDTO: Codable, Hashable, Identifiable {
    let accountNum: Int
    let email: String
    let displayName: String
    let isActive: Bool
    let checks: [CheckResultDTO]
    let swapReady: Bool

    var id: Int { accountNum }
}

/// Mirrors VerificationReport.
struct VerificationReportDTO: Codable, Hashable {
    let results: [AccountVerificationDTO]
    let total: Int
    let ready: Int
    let failed: Int
}
