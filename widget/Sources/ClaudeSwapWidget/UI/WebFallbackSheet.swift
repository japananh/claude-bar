import SwiftUI
import WebKit

/// Floating-window content for the claude.ai web fallback.
///
/// Layout: address row + WebView + status footer.
/// User logs in once → cookies persist → cản be reused next launch.
struct WebFallbackSheet: View {
    @EnvironmentObject var coordinator: WebFallbackCoordinator

    @State private var currentURL: URL? = URL(string: "https://claude.ai/")
    @State private var isLoading = false
    @State private var pageTitle = ""
    @State private var scrapeResult: String?

    private let homeURL = URL(string: "https://claude.ai/settings/limits")!

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            addressRow
            Divider()
            ClaudeWebView(
                initialURL: homeURL,
                currentURL: $currentURL,
                isLoading: $isLoading,
                title: $pageTitle,
                onCookiesChanged: {
                    Task { await coordinator.refreshConnectionState() }
                }
            )
            Divider()
            footer
        }
        .frame(width: 720, height: 640)
    }

    private var addressRow: some View {
        HStack(spacing: 8) {
            Image(systemName: "lock.fill")
                .foregroundColor(.green).font(.system(size: 10))
            Text(currentURL?.absoluteString ?? "")
                .font(.system(size: 11, design: .monospaced))
                .lineLimit(1).truncationMode(.middle)
            Spacer()
            if isLoading {
                ProgressView().controlSize(.mini)
            }
        }
        .padding(.horizontal, 12).padding(.vertical, 6)
    }

    private var footer: some View {
        HStack(spacing: 8) {
            connectionBadge
            Spacer()
            if let txt = scrapeResult {
                Text(txt)
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
                    .lineLimit(1).truncationMode(.middle)
                    .help(txt)
            }
            Button("Try to scrape quota") {
                Task { await scrapeQuota() }
            }
            .controlSize(.small)
            Button("Disconnect", role: .destructive) {
                Task {
                    await coordinator.disconnect()
                    scrapeResult = nil
                }
            }
            .controlSize(.small)
            .disabled(!coordinator.isConnected)
            Button("Close") { coordinator.dismiss() }
                .controlSize(.small)
                .keyboardShortcut(.defaultAction)
        }
        .padding(.horizontal, 12).padding(.vertical, 8)
    }

    private var connectionBadge: some View {
        HStack(spacing: 4) {
            Image(systemName: coordinator.isConnected
                  ? "checkmark.shield.fill"
                  : "lock.slash.fill")
                .foregroundColor(coordinator.isConnected ? .green : .orange)
                .font(.system(size: 11))
            Text(coordinator.isConnected ? "Connected" : "Not signed in")
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
    }

    /// Best-effort scrape: read the page text and try to find quota-shaped
    /// strings. Anthropic's HTML is private so the JS is intentionally loose.
    private func scrapeQuota() async {
        let js = """
        (function () {
          const txt = document.body.innerText || "";
          const lines = txt.split('\\n').map(s => s.trim()).filter(Boolean);
          const pat = /(\\d+(\\.\\d+)?\\s*%)|(\\d+\\s+(messages?|of)\\s+\\d+)/i;
          for (const l of lines) { if (pat.test(l)) return l.slice(0, 200); }
          return null;
        })();
        """
        guard let webView = findWebView() else { return }
        do {
            let val = try await webView.evaluateJavaScript(js)
            scrapeResult = (val as? String) ?? "No quota text found on this page"
            coordinator.lastScrapedQuotaText = scrapeResult
        } catch {
            scrapeResult = "Scrape failed: \(error.localizedDescription)"
        }
    }

    /// Walk the app's window hierarchy to find the WKWebView NSViewRepresentable
    /// renders. Simple traversal — there is at most one in this floating window.
    private func findWebView() -> WKWebView? {
        for window in NSApp.windows {
            if let v = findWebView(in: window.contentView) { return v }
        }
        return nil
    }

    private func findWebView(in view: NSView?) -> WKWebView? {
        guard let view else { return nil }
        if let w = view as? WKWebView { return w }
        for sub in view.subviews {
            if let w = findWebView(in: sub) { return w }
        }
        return nil
    }
}
