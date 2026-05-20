import SwiftUI
import WebKit

/// NSViewRepresentable wrapper around WKWebView, the same engine Safari uses.
///
/// Configured with the **default** WKWebsiteDataStore so cookies + localStorage
/// persist across widget launches. This lets the user log in to claude.ai
/// once and reuse the session as a backup when the OAuth usage API is
/// rate-limited at the Cloudflare layer.
struct ClaudeWebView: NSViewRepresentable {
    let initialURL: URL
    @Binding var currentURL: URL?
    @Binding var isLoading: Bool
    @Binding var title: String
    let onCookiesChanged: () -> Void

    func makeNSView(context: Context) -> WKWebView {
        let config = WKWebViewConfiguration()
        config.websiteDataStore = .default() // persistent, shared with system
        let view = WKWebView(frame: .zero, configuration: config)
        view.navigationDelegate = context.coordinator
        view.uiDelegate = context.coordinator
        view.load(URLRequest(url: initialURL))
        return view
    }

    func updateNSView(_: WKWebView, context _: Context) {}

    func makeCoordinator() -> Coordinator {
        Coordinator(parent: self)
    }

    final class Coordinator: NSObject, WKNavigationDelegate, WKUIDelegate {
        let parent: ClaudeWebView
        init(parent: ClaudeWebView) { self.parent = parent }

        func webView(_ webView: WKWebView, didStartProvisionalNavigation _: WKNavigation!) {
            parent.isLoading = true
        }

        func webView(_ webView: WKWebView, didFinish _: WKNavigation!) {
            parent.isLoading = false
            parent.currentURL = webView.url
            parent.title = webView.title ?? ""
            parent.onCookiesChanged()
        }

        func webView(_: WKWebView, didFail _: WKNavigation!, withError _: Error) {
            parent.isLoading = false
        }
    }
}

/// Reads claude.ai cookies from the shared WKWebsiteDataStore.
enum ClaudeWebSession {
    /// Returns true if a session-looking cookie exists for claude.ai.
    static func isConnected() async -> Bool {
        let store = WKWebsiteDataStore.default().httpCookieStore
        let cookies = await store.allCookies()
        return cookies.contains {
            $0.domain.hasSuffix("claude.ai") &&
            ($0.name.lowercased().contains("session") || $0.name.lowercased().contains("auth"))
        }
    }

    /// Returns (cookieName, value) pairs for claude.ai (diagnostics).
    static func sessionSummary() async -> [String: String] {
        let cookies = await WKWebsiteDataStore.default().httpCookieStore.allCookies()
        var out: [String: String] = [:]
        for c in cookies where c.domain.hasSuffix("claude.ai") {
            out[c.name] = "\(c.value.prefix(8))…"
        }
        return out
    }

    /// Clears all claude.ai cookies (logout from web fallback).
    static func clear() async {
        let store = WKWebsiteDataStore.default().httpCookieStore
        let cookies = await store.allCookies()
        for c in cookies where c.domain.hasSuffix("claude.ai") {
            await store.deleteCookie(c)
        }
    }
}
