// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "ClaudeSwapWidget",
    platforms: [.macOS(.v14)],
    products: [
        .executable(name: "ClaudeSwapWidget", targets: ["ClaudeSwapWidget"])
    ],
    targets: [
        .executableTarget(
            name: "ClaudeSwapWidget",
            path: "Sources/ClaudeSwapWidget",
            resources: [.process("Resources")]
        ),
        .testTarget(
            name: "ClaudeSwapWidgetTests",
            dependencies: ["ClaudeSwapWidget"],
            path: "Tests/ClaudeSwapWidgetTests"
        )
    ]
)
