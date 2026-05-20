import SwiftUI

struct RenameAccountSheet: View {
    let account: AccountViewDTO
    let onSubmit: (String) -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var nickname: String = ""

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Rename profile").font(.headline)
            Text(account.account.email).font(.caption).foregroundColor(.secondary)

            TextField("Profile name", text: $nickname)
                .textFieldStyle(.roundedBorder)
                .onSubmit(submit)

            HStack {
                Button("Clear") {
                    onSubmit("")
                    dismiss()
                }
                .disabled((account.account.nickname ?? "").isEmpty)
                Spacer()
                Button("Cancel") { dismiss() }
                    .keyboardShortcut(.escape, modifiers: [])
                Button("Save", action: submit)
                    .keyboardShortcut(.defaultAction)
                    .disabled(nickname.trimmingCharacters(in: .whitespaces).isEmpty)
            }
        }
        .padding(20)
        .frame(width: 320)
        .onAppear { nickname = account.account.nickname ?? "" }
    }

    private func submit() {
        let trimmed = nickname.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return }
        onSubmit(trimmed)
        dismiss()
    }
}
