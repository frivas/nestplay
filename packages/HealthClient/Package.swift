// swift-tools-version: 6.0
import PackageDescription

let package = Package(
  name: "HealthClient",
  platforms: [.macOS(.v14), .tvOS(.v17)],
  products: [
    .library(name: "HealthClient", targets: ["HealthClient"]),
    .executable(name: "health-probe", targets: ["HealthProbe"]),
  ],
  targets: [
    .target(name: "HealthClient"),
    .executableTarget(name: "HealthProbe", dependencies: ["HealthClient"]),
    .testTarget(name: "HealthClientTests", dependencies: ["HealthClient"]),
  ]
)
