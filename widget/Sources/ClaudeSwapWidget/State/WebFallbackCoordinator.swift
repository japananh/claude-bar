import Foundation
import SwiftUI
import AppKit

/// Owns the floating window that hosts the embedded claude.ai WebView.
///
/// Tracks whether the user has a live web session (cookies present) so the
/// menu UI can show "web fallback available" when the OAuth usage API is
/// rate-limited.
@MainActor
final class WebFallbackCoordinator: ObservableObject {
    @Published var isConnected: Bool = false
    @Published var lastCheckedAt: Date?
    @Published var lastScrapedQuotaText: String?

    private let window = FloatingWindow<AnyView>()
    private weak var store: AppStore?

    func attach(store: AppStore) {
        self.store = store
        Task { await refreshConnectionState() }
    }

    /// Open the floating window with the embedded claude.ai browser.
    func open() {
        guard let store else { return }
        window.show(title: "Claude.ai — web fallback",
                    size: NSSize(width: 720, height: 640)) {
            AnyView(
                WebFallbackSheet()
                    .environmentObject(store)
                    .environmentObject(self)
            )
        }
    }

    func dismiss() {
        window.close()
    }

    /// Probe cookies to decide whether the user is signed in to claude.ai.
    func refreshConnectionState() async {
        isConnected = await ClaudeWebSession.isConnected()
        lastCheckedAt = Date()
    }

    /// Clear claude.ai cookies. User will need to log in again next open.
    func disconnect() async {
        await ClaudeWebSession.clear()
        lastScrapedQuotaText = nil
        await refreshConnectionState()
    }
}
