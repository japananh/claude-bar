import SwiftUI

/// Circular avatar with a single letter and a stable color derived from
/// the account's identity. Helps users distinguish profiles at a glance
/// when the list grows past 2-3 accounts.
struct AvatarView: View {
    let initial: String
    let seed: String
    var size: CGFloat = 24

    var body: some View {
        ZStack {
            Circle().fill(
                LinearGradient(
                    colors: [color, color.opacity(0.7)],
                    startPoint: .topLeading,
                    endPoint: .bottomTrailing
                )
            )
            Text(initial)
                .font(.system(size: size * 0.45, weight: .semibold, design: .rounded))
                .foregroundColor(.white)
        }
        .frame(width: size, height: size)
    }

    private var color: Color {
        Self.palette[abs(stableHash(seed)) % Self.palette.count]
    }

    private static let palette: [Color] = [
        Color(red: 0.20, green: 0.55, blue: 0.95),  // blue
        Color(red: 0.95, green: 0.45, blue: 0.30),  // coral
        Color(red: 0.30, green: 0.75, blue: 0.45),  // green
        Color(red: 0.75, green: 0.40, blue: 0.85),  // purple
        Color(red: 0.95, green: 0.65, blue: 0.20),  // amber
        Color(red: 0.20, green: 0.70, blue: 0.75),  // teal
        Color(red: 0.85, green: 0.30, blue: 0.55),  // pink
        Color(red: 0.45, green: 0.50, blue: 0.95),  // indigo
    ]

    private func stableHash(_ s: String) -> Int {
        var h = 0
        for u in s.unicodeScalars { h = h &* 31 &+ Int(u.value) }
        return h
    }
}
