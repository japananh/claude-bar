import XCTest
@testable import ClaudeSwapWidget

final class SmokeTests: XCTestCase {
    func testAccountDisplayNameFallsBackToEmail() {
        let acc = AccountDTO(
            number: 1, email: "alice@example.com",
            organizationName: nil, organizationUuid: nil,
            nickname: nil, createdAt: Date()
        )
        XCTAssertEqual(acc.displayName, "alice@example.com")
    }

    func testAccountDisplayNameUsesNickname() {
        let acc = AccountDTO(
            number: 1, email: "alice@example.com",
            organizationName: nil, organizationUuid: nil,
            nickname: "Personal", createdAt: Date()
        )
        XCTAssertEqual(acc.displayName, "Personal")
    }
}
