import SwiftUI

/// Small-caps section label, macOS native menu style.
/// Optionally renders a trailing detail (e.g. row count).
struct SectionHeaderView: View {
    let title: String
    var trailing: String? = nil
    var color: Color = .secondary

    var body: some View {
        HStack {
            Text(title.uppercased())
                .font(.system(size: 10, weight: .semibold))
                .foregroundColor(color)
                .tracking(0.6)
            Spacer()
            if let trailing {
                Text(trailing)
                    .font(.system(size: 10, weight: .medium))
                    .foregroundColor(color)
                    .monospacedDigit()
            }
        }
        .padding(.horizontal, 12)
        .padding(.top, 8)
        .padding(.bottom, 2)
    }
}
