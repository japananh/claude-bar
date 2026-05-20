import Foundation
import SwiftUI
import AppKit

/// Drives the "Verify all accounts" diagnostic flow.
///
/// Opens a floating window, runs `csw verify --json` in the background,
/// streams the result back to the UI.
@MainActor
final class VerifyCoordinator: ObservableObject {
    enum Phase: Equatable {
        case idle
        case running
        case done(VerificationReportDTO)
        case failed(String)
    }

    @Published var phase: Phase = .idle

    private let window = FloatingWindow<AnyView>()
    private weak var store: AppStore?

    func attach(store: AppStore) { self.store = store }

    func begin() {
        guard let store else { return }
        phase = .idle
        window.show(title: "Verify accounts", size: NSSize(width: 480, height: 440)) {
            AnyView(
                VerifyAccountsSheet()
                    .environmentObject(store)
                    .environmentObject(self)
            )
        }
        Task { await run(client: store.client) }
    }

    func run(client: CswClient) async {
        phase = .running
        do {
            let report = try await client.verify()
            phase = .done(report)
        } catch {
            phase = .failed(error.localizedDescription)
        }
    }

    func dismiss() {
        window.close()
        phase = .idle
    }
}
