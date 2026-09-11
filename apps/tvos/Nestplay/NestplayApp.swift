import SwiftUI

@main
struct NestplayApp: App {
  var body: some Scene {
    WindowGroup { ConnectionView().frame(maxWidth: .infinity, maxHeight: .infinity) }
  }
}
