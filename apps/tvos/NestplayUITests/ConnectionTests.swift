import XCTest

final class ConnectionTests: XCTestCase {
  @MainActor
  func testConnectsToRealGoServer() {
    let app = XCUIApplication()
    app.launchArguments = ["--server-url", "http://127.0.0.1:8765"]
    app.launch()
    XCTAssertTrue(app.staticTexts["Connected to Nestplay"].waitForExistence(timeout: 15))
  }

  @MainActor
  func testUnavailableServerAllowsRetry() {
    let app = XCUIApplication()
    app.launchArguments = ["--server-url", "http://127.0.0.1:1"]
    app.launch()
    XCTAssertTrue(app.staticTexts["Connection needs attention"].waitForExistence(timeout: 15))
    XCTAssertTrue(app.buttons["checkConnection"].isEnabled)
  }
}
