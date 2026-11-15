// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "MoneyRecapContract",
    platforms: [.iOS(.v17), .macOS(.v14)],
    products: [
        .library(name: "MoneyRecapContract", targets: ["MoneyRecapContract"]),
    ],
    dependencies: [
        .package(url: "https://github.com/connectrpc/connect-swift.git", exact: "1.2.3"),
        .package(url: "https://github.com/apple/swift-protobuf.git", exact: "1.39.0"),
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
