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

  @MainActor
  func testLaunchOpensLocalLibraryWithYouTubeHomeSwitch() {
    let app = XCUIApplication()
    app.launchArguments = ["--server-url", "http://127.0.0.1:8765"]
    app.launch()
    XCTAssertTrue(app.staticTexts["Connected to Nestplay"].waitForExistence(timeout: 15))
    XCTAssertTrue(app.otherElements["localLibrarySurface"].waitForExistence(timeout: 5))
    XCTAssertTrue(
      app.otherElements["libraryCatalog"].exists || app.staticTexts["libraryEmpty"].exists)
    XCTAssertFalse(app.otherElements["youtubeHomeSurface"].isHittable)
    let youtubeHome = youtubeHomeSwitch(in: app)
    XCTAssertTrue(youtubeHome.waitForExistence(timeout: 5))
    youtubeHome.tap()
    XCTAssertTrue(app.otherElements["youtubeHomeSurface"].waitForExistence(timeout: 5))
    XCTAssertTrue(app.staticTexts["youtubeHomeEmpty"].exists)
  }

  private func youtubeHomeSwitch(in app: XCUIApplication) -> XCUIElement {
    let identified = app.descendants(matching: .any)["youtubeHomeTab"]
    if identified.exists { return identified }
    if app.buttons["YouTube Home"].exists { return app.buttons["YouTube Home"] }
    return app.tabBars.buttons["YouTube Home"]
  }
}
