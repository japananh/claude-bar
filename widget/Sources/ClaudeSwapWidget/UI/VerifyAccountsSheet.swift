import SwiftUI

/// Result panel for the "Verify accounts" diagnostic.
/// Lists each account and the outcome of each check (passed / skipped / failed).
struct VerifyAccountsSheet: View {
    @EnvironmentObject var store: AppStore
    @EnvironmentObject var coordinator: VerifyCoordinator

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            header
            Divider()
            content
            Spacer(minLength: 0)
            Divider()
            footer
        }
        .padding(20)
        .frame(width: 480, height: 440)
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text("Verify accounts").font(.title2).fontWeight(.semibold)
            Text("Checks whether every account can be swapped to successfully.")
                .font(.subheadline).foregroundColor(.secondary)
        }
    }

    @ViewBuilder
    private var content: some View {
        switch coordinator.phase {
        case .idle, .running:
            VStack(spacing: 12) {
                ProgressView()
                Text("Probing each account…")
                    .font(.callout).foregroundColor(.secondary)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        case .done(let report):
            VerifyReportView(report: report)
        case .failed(let msg):
            VStack(alignment: .leading, spacing: 8) {
                Label(msg, systemImage: "xmark.octagon.fill")
                    .foregroundColor(.red)
                Text("Run `csw verify` in a terminal for more detail.")
                    .font(.caption).foregroundColor(.secondary)
            }
        }
    }

    @ViewBuilder
    private var footer: some View {
        HStack {
            if case .done = coordinator.phase {
                Button("Re-run") {
                    Task { await coordinator.run(client: store.client) }
                }
            }
            Spacer()
            Button("Close") { coordinator.dismiss() }
                .keyboardShortcut(.defaultAction)
        }
    }
}

private struct VerifyReportView: View {
    let report: VerificationReportDTO

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            summaryHeader
            ScrollView {
                VStack(alignment: .leading, spacing: 10) {
                    ForEach(report.results) { res in
                        VerifyAccountRow(verification: res)
                    }
                }
            }
        }
    }

    private var summaryHeader: some View {
        HStack(spacing: 8) {
            Image(systemName: report.failed == 0
                  ? "checkmark.shield.fill"
                  : "exclamationmark.shield.fill")
                .foregroundColor(report.failed == 0 ? .green : .orange)
            Text("\(report.ready) of \(report.total) account\(report.total == 1 ? "" : "s") swap-ready")
                .font(.callout).fontWeight(.semibold)
            Spacer()
        }
        .padding(.horizontal, 6)
    }
}

private struct VerifyAccountRow: View {
    let verification: AccountVerificationDTO

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 8) {
                AvatarView(initial: String(verification.displayName.prefix(1)).uppercased(),
                           seed: verification.email, size: 22)
                Text(verification.displayName).font(.system(size: 13, weight: .semibold))
                Text(verification.email).font(.caption).foregroundColor(.secondary)
                Spacer()
                if verification.swapReady {
                    badge(text: "READY", color: .green)
                } else {
                    badge(text: "FAILED", color: .red)
                }
            }
            ForEach(verification.checks) { c in
                checkRow(c)
            }
            if !verification.swapReady {
                Text("→ Re-add this account via the menu (\"+ Add account\").")
                    .font(.caption2).foregroundColor(.orange)
                    .padding(.leading, 24)
            }
        }
        .padding(10)
        .background(Color.primary.opacity(0.05))
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }

    private func checkRow(_ c: CheckResultDTO) -> some View {
        HStack(alignment: .top, spacing: 6) {
            Image(systemName: glyph(c))
                .foregroundColor(color(c))
                .font(.system(size: 11))
                .frame(width: 16)
            VStack(alignment: .leading, spacing: 1) {
                Text(c.label).font(.caption)
                if let detail = c.detail, !detail.isEmpty {
                    Text(detail).font(.caption2).foregroundColor(.secondary)
                        .lineLimit(2).truncationMode(.middle)
                }
            }
        }
        .padding(.leading, 22)
    }

    private func glyph(_ c: CheckResultDTO) -> String {
        if c.skipped == true { return "minus.circle" }
        return c.passed ? "checkmark.circle.fill" : "xmark.circle.fill"
    }

    private func color(_ c: CheckResultDTO) -> Color {
        if c.skipped == true { return .secondary }
        return c.passed ? .green : .red
    }

    private func badge(text: String, color: Color) -> some View {
        Text(text)
            .font(.system(size: 9, weight: .bold))
            .tracking(0.5)
            .foregroundColor(.white)
            .padding(.horizontal, 6).padding(.vertical, 2)
            .background(color)
            .clipShape(Capsule())
    }
}
