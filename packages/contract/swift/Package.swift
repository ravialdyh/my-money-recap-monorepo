// swift-tools-version: 5.9
import PackageDescription

/// The connect-swift client generated from packages/contract by `buf generate`
/// (ARCHITECTURE A1, IOS I3). Nothing in Sources/ is written by hand.
let package = Package(
    name: "MoneyRecapContract",
    platforms: [.iOS(.v17), .macOS(.v14)],
    products: [
        .library(name: "MoneyRecapContract", targets: ["MoneyRecapContract"]),
    ],
    dependencies: [
        .package(url: "https://github.com/connectrpc/connect-swift.git", exact: "1.2.3"),
        .package(url: "https://github.com/apple/swift-protobuf.git", exact: "1.38.1"),
    ],
    targets: [
        .target(
            name: "MoneyRecapContract",
            dependencies: [
                .product(name: "Connect", package: "connect-swift"),
                .product(name: "SwiftProtobuf", package: "swift-protobuf"),
            ]
        ),
    ]
)
