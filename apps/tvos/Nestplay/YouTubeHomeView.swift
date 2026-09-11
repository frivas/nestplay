import SwiftUI

struct YouTubeHomeView: View {
  var body: some View {
    VStack(alignment: .leading, spacing: 24) {
      Text("YouTube Home")
        .font(.system(size: 56, weight: .semibold))
      Text("YouTube Home is not loaded yet.")
        .font(.title3)
        .foregroundStyle(.secondary)
        .accessibilityIdentifier("youtubeHomeEmpty")
    }
    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
    .padding(60)
    .accessibilityIdentifier("youtubeHomeSurface")
  }
}
