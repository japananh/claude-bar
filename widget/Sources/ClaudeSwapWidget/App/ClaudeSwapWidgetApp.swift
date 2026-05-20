import SwiftUI
import UserNotifications

extension Notification.Name {
    static let openSettings = Notification.Name("claudebar.openSettings")
}

@main
struct ClaudeSwapWidgetApp: App {
    @StateObject private var store = AppStore()
    @StateObject private var loginCoordinator = LoginCoordinator()
    @StateObject private var verifyCoordinator = VerifyCoordinator()
    @StateObject private var webFallback = WebFallbackCoordinator()
    @ObservedObject private var settings = AppSettings.shared
    @Environment(\.openSettings) private var openSettings

    init() {
        UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound]) { _, _ in }
        ClaudeWatchInstaller.install()
    }

    var body: some Scene {
        MenuBarExtra {
            MenuContentView()
                .environmentObject(store)
                .environmentObject(loginCoordinator)
                .environmentObject(verifyCoordinator)
                .environmentObject(webFallback)
                .frame(width: 400)
                .task {
                    loginCoordinator.attach(store: store)
                    verifyCoordinator.attach(store: store)
                    webFallback.attach(store: store)
                    store.start()
                }
                .onReceive(NotificationCenter.default.publisher(for: .openSettings)) { _ in
                    openSettings()
                    let abovePopup = NSWindow.Level(rawValue: NSWindow.Level.statusBar.rawValue + 1)
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.15) {
                        NSApp.activate(ignoringOtherApps: true)
                        NSApp.windows
                            .filter { $0.isVisible && $0.canBecomeKey }
                            .filter { $0.level == .normal }
                            .forEach {
                                $0.level = abovePopup
                                $0.makeKeyAndOrderFront(nil)
                            }
                    }
                }
        } label: {
            MenuBarLabelView()
                .environmentObject(store)
        }
        .menuBarExtraStyle(.window)

        Settings {
            SettingsWindowView()
                .environmentObject(store)
                .environmentObject(loginCoordinator)
                .environmentObject(verifyCoordinator)
                .environmentObject(webFallback)
        }
    }
}
